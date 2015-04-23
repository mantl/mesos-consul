package zkdetect

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"sync"
	"testing"
	"time"

	log "github.com/golang/glog"
	"github.com/samuel/go-zookeeper/zk"
	"github.com/stretchr/testify/assert"
)

const (
	zkurl     = "zk://127.0.0.1:2181/mesos"
	zkurl_bad = "zk://127.0.0.1:2181"
)

func TestParseZk_single(t *testing.T) {
	hosts, path, err := parseZk(zkurl)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(hosts))
	assert.Equal(t, "/mesos", path)
}

func TestParseZk_multi(t *testing.T) {
	hosts, path, err := parseZk("zk://abc:1,def:2/foo")
	assert.NoError(t, err)
	assert.Equal(t, []string{"abc:1", "def:2"}, hosts)
	assert.Equal(t, "/foo", path)
}

func TestParseZk_multiIP(t *testing.T) {
	hosts, path, err := parseZk("zk://10.186.175.156:2181,10.47.50.94:2181,10.0.92.171:2181/mesos")
	assert.NoError(t, err)
	assert.Equal(t, []string{"10.186.175.156:2181", "10.47.50.94:2181", "10.0.92.171:2181"}, hosts)
	assert.Equal(t, "/mesos", path)
}

func TestClusterDetectorStart(t *testing.T) {
	c, err := makeClient()
	assert.False(t, c.isConnected())
	md, err := NewClusterDetector(zkurl)
	defer md.Cancel()
	assert.NoError(t, err)
	c.errorHandler = ErrorHandler(func(c *Client, e error) {
		err = e
	})
	md.client = c // override zk.Conn with our own.
	md.client.connect()
	assert.NoError(t, err)
	assert.True(t, c.isConnected())
}

func TestClusterDetectorChildrenChanged(t *testing.T) {
	wCh := make(chan struct{}, 1)

	c, err := makeClient()
	assert.NoError(t, err)
	assert.False(t, c.isConnected())

	md, err := NewClusterDetector(zkurl)
	defer md.Cancel()
	assert.NoError(t, err)
	// override zk.Conn with our own.
	c.errorHandler = ErrorHandler(func(c *Client, e error) {
		err = e
	})
	md.client = c
	md.client.connect()
	assert.NoError(t, err)
	assert.True(t, c.isConnected())

	called := 0
	md.Detect(OnClusterChanged(func(cluster *ClusterInfo) {
		//expect 2 calls in sequence: the first setting a master
		//and the second clearing it
		switch called++; called {
		case 1:
			assert.NotNil(t, cluster.leader)
			assert.Equal(t, cluster.leader.GetId(), "master@localhost:5050")
			wCh <- struct{}{}
		case 2:
			assert.Equal(t, cluster.leader.GetId(), "master@localhost:5050")
			//assert.Nil(t, cluster.leader)
			wCh <- struct{}{}
		case 3:
			assert.Nil(t, cluster)
			wCh <- struct{}{}
		default:
			t.Fatalf("unexpected notification call attempt %d", called)
		}
	}))

	startWait := time.Now()
	select {
	case <-wCh:
	case <-time.After(time.Second * 3):
		panic("Waited too long...")
	}

	// wait for the disconnect event, should be triggered
	// 1s after the connected event
	waited := time.Now().Sub(startWait)
	time.Sleep((2 * time.Second) - waited)
	assert.False(t, c.isConnected())
}

// single connector instance, session does not expire, but it's internal connection to zk is flappy
func TestClusterDetectFlappingConnectionState(t *testing.T) {
	c, err := newClient(test_zk_hosts, test_zk_path)
	assert.NoError(t, err)

	initialChildren := []string{"info_005", "info_010", "info_022"}
	connector := NewMockConnector()
	connector.On("Close").Return(nil)
	connector.On("Children", test_zk_path).Return(initialChildren, &zk.Stat{}, nil)

	var wg sync.WaitGroup
	wg.Add(2) // async flapping, master change detection

	first := true
	c.setFactory(asFactory(func() (Connector, <-chan zk.Event, error) {
		if !first {
			t.Fatalf("only one connector instance expected")
			return nil, nil, errors.New("ran out of connectors")
		} else {
			first = false
		}
		sessionEvents := make(chan zk.Event, 10)
		watchEvents := make(chan zk.Event, 10)

		connector.On("Get", fmt.Sprintf("%s/info_005", test_zk_path)).Return(newTestClusterInfo("info_005"),
			&zk.Stat{}, nil).Once()
		connector.On("Get", fmt.Sprintf("%s/info_010", test_zk_path)).Return(newTestClusterInfo("info_010"),
			&zk.Stat{}, nil).Once()
		connector.On("Get", fmt.Sprintf("%s/info_022", test_zk_path)).Return(newTestClusterInfo("info_022"),
			&zk.Stat{}, nil).Once()
		connector.On("ChildrenW", test_zk_path).Return([]string{test_zk_path}, &zk.Stat{}, (<-chan zk.Event)(watchEvents), nil).Once()
		go func() {
			defer wg.Done()
			time.Sleep(100 * time.Millisecond)
			for attempt := 0; attempt < 5; attempt++ {
				sessionEvents <- zk.Event{
					Type:  zk.EventSession,
					State: zk.StateConnected,
				}
				time.Sleep(500 * time.Millisecond)
				sessionEvents <- zk.Event{
					Type:  zk.EventSession,
					State: zk.StateDisconnected,
				}
			}
			sessionEvents <- zk.Event{
				Type:  zk.EventSession,
				State: zk.StateConnected,
			}
		}()
		return connector, sessionEvents, nil
	}))
	c.reconnDelay = 0 // there should be no reconnect, but just in case don't drag the test out

	md, err := NewClusterDetector(zkurl)
	defer md.Cancel()
	assert.NoError(t, err)

	c.errorHandler = ErrorHandler(func(c *Client, e error) {
		t.Logf("zk client error: %v", e)
	})
	md.client = c

	startTime := time.Now()
	detected := false
	md.Detect(OnClusterChanged(func(cluster *ClusterInfo) {
		if detected {
			t.Fatalf("already detected master, was not expecting another change: %v", cluster)
		} else {
			detected = true
			assert.NotNil(t, cluster.leader, fmt.Sprintf("on-master-changed %v", detected))
			t.Logf("Leader change detected at %v: '%+v'", time.Now().Sub(startTime), cluster.leader)
			wg.Done()
		}
	}))

	completed := make(chan struct{})
	go func() {
		defer close(completed)
		wg.Wait()
	}()

	select {
	case <-completed: // expected
	case <-time.After(3 * time.Second):
		t.Fatalf("failed to detect master change")
	}
}

func TestClusterDetectFlappingConnector(t *testing.T) {
	c, err := newClient(test_zk_hosts, test_zk_path)
	assert.NoError(t, err)

	initialChildren := []string{"info_005", "info_010", "info_022"}
	connector := NewMockConnector()
	connector.On("Close").Return(nil)
	connector.On("Children", test_zk_path).Return(initialChildren, &zk.Stat{}, nil)

	attempt := 0
	c.setFactory(asFactory(func() (Connector, <-chan zk.Event, error) {
		attempt++
		sessionEvents := make(chan zk.Event, 5)
		watchEvents := make(chan zk.Event, 5)

		sessionEvents <- zk.Event{
			Type:  zk.EventSession,
			State: zk.StateConnected,
		}
		for _, id := range initialChildren {
			connector.On("Get", fmt.Sprintf("%s/%s", test_zk_path, id)).Return(newTestClusterInfo(id),
				&zk.Stat{}, nil).Once()
		}
		connector.On("ChildrenW", test_zk_path).Return([]string{test_zk_path}, &zk.Stat{}, (<-chan zk.Event)(watchEvents), nil).Once()
		go func(attempt int) {
			defer close(sessionEvents)
			defer close(watchEvents)
			time.Sleep(400 * time.Millisecond)
			// this is the order in which the embedded zk implementation does it
			sessionEvents <- zk.Event{
				Type:  zk.EventSession,
				State: zk.StateDisconnected,
			}
			connector.On("ChildrenW", test_zk_path).Return(nil, nil, nil, zk.ErrSessionExpired).Once()
			watchEvents <- zk.Event{
				Type:  zk.EventNotWatching,
				State: zk.StateDisconnected,
				Path:  test_zk_path,
				Err:   zk.ErrSessionExpired,
			}
		}(attempt)
		return connector, sessionEvents, nil
	}))
	c.reconnDelay = 100 * time.Millisecond

	md, err := NewClusterDetector(zkurl)
	defer md.Cancel()
	assert.NoError(t, err)

	c.errorHandler = ErrorHandler(func(c *Client, e error) {
		t.Logf("zk client error: %v", e)
	})
	md.client = c

	var wg sync.WaitGroup
	wg.Add(4) // 2 x (connected, disconnected)
	detected := 0
	startTime := time.Now()
	md.Detect(OnClusterChanged(func(cluster *ClusterInfo) {
		if detected > 3 {
			// ignore
			return
		}
		if (detected & 1) == 0 {
			assert.NotNil(t, cluster.leader, fmt.Sprintf("on-master-changed-%d", detected))
		} else {
			assert.Nil(t, cluster, fmt.Sprintf("on-master-changed-%d", detected))
		}
		t.Logf("Leader change detected at %v: '%+v'", time.Now().Sub(startTime), cluster)
		detected++
		wg.Done()
	}))

	completed := make(chan struct{})
	go func() {
		defer close(completed)
		wg.Wait()
	}()

	select {
	case <-completed: // expected
	case <-time.After(3 * time.Second):
		t.Fatalf("failed to detect flapping master changes")
	}
}

func TestClusterDetectMultiple(t *testing.T) {
	ch0 := make(chan zk.Event, 5)
	ch1 := make(chan zk.Event, 5)

	ch0 <- zk.Event{
		Type:  zk.EventSession,
		State: zk.StateConnected,
	}

	c, err := newClient(test_zk_hosts, test_zk_path)
	assert.NoError(t, err)

	initialChildren := []string{"info_005", "info_010", "info_022"}
	connector := NewMockConnector()
	connector.On("Close").Return(nil)
	connector.On("Children", test_zk_path).Return(initialChildren, &zk.Stat{}, nil).Once()
	connector.On("ChildrenW", test_zk_path).Return([]string{test_zk_path}, &zk.Stat{}, (<-chan zk.Event)(ch1), nil)
	for _, id := range initialChildren {
		connector.On("Get", fmt.Sprintf("%s/%s", test_zk_path, id)).Return(newTestClusterInfo(id),
			&zk.Stat{}, nil).Once()
	}

	first := true
	c.setFactory(asFactory(func() (Connector, <-chan zk.Event, error) {
		log.Print("[INFO] **** Using zk.Conn adapter ****")
		if !first {
			return nil, nil, errors.New("only 1 connector allowed")
		} else {
			first = false
		}
		return connector, ch0, nil
	}))

	md, err := NewClusterDetector(zkurl)
	defer md.Cancel()
	assert.NoError(t, err)

	c.errorHandler = ErrorHandler(func(c *Client, e error) {
		err = e
	})
	md.client = c

	// **** Test 4 consecutive ChildrenChangedEvents ******
	// setup event changes
	sequences := [][]string{
		[]string{"info_014", "info_010", "info_005"},
		[]string{"info_005", "info_004", "info_022"},
		[]string{}, // indicates no master
		[]string{"info_017", "info_099", "info_200"},
	}

	var wg sync.WaitGroup
	startTime := time.Now()
	detected := 0
	md.Detect(OnClusterChanged(func(cluster *ClusterInfo) {
		if detected == 3 {
			assert.Nil(t, cluster.leader, fmt.Sprintf("on-master-changed-%d", detected))
		} else {
			assert.NotNil(t, cluster.leader, fmt.Sprintf("on-master-changed-%d", detected))
		}
		t.Logf("Leader change detected at %v: '%+v'", time.Now().Sub(startTime), cluster.leader)
		detected++
		wg.Done()
	}))

	// 3 leadership changes + disconnect (leader change to '')
	wg.Add(5)

	go func() {
		for i := range sequences {
			sorted := make([]string, len(sequences[i]))
			copy(sorted, sequences[i])
			sort.Strings(sorted)
			t.Logf("testing master change sequence %d, path '%v'", i, test_zk_path)
			connector.On("Children", test_zk_path).Return(sequences[i], &zk.Stat{}, nil).Once()
			if len(sequences[i]) > 0 {
				connector.On("Get", fmt.Sprintf("%s/%s", test_zk_path, sorted[0])).Return(newTestClusterInfo(sorted[0]), &zk.Stat{}, nil).Once()
				connector.On("Get", fmt.Sprintf("%s/%s", test_zk_path, sorted[1])).Return(newTestClusterInfo(sorted[1]), &zk.Stat{}, nil).Once()
				connector.On("Get", fmt.Sprintf("%s/%s", test_zk_path, sorted[2])).Return(newTestClusterInfo(sorted[2]), &zk.Stat{}, nil).Once()
			}
			ch1 <- zk.Event{
				Type: zk.EventNodeChildrenChanged,
				Path: test_zk_path,
			}
			time.Sleep(100 * time.Millisecond) // give async routines time to catch up
		}
		time.Sleep(1 * time.Second) // give async routines time to catch up
		t.Logf("disconnecting...")
		ch0 <- zk.Event{
			State: zk.StateDisconnected,
		}
		//TODO(jdef) does order of close matter here? probably, meaking client code is weak
		close(ch0)
		time.Sleep(500 * time.Millisecond) // give async routines time to catch up
		close(ch1)
	}()
	completed := make(chan struct{})
	go func() {
		defer close(completed)
		wg.Wait()
	}()

	defer func() {
		if r := recover(); r != nil {
			t.Fatal(r)
		}
	}()

	select {
	case <-time.After(2 * time.Second):
		panic("timed out waiting for master changes to propagate")
	case <-completed:
	}
}

func TestClusterDetect_selectTopNode_none(t *testing.T) {
	assert := assert.New(t)
	nodeList := []string{}
	node := selectTopNode(nodeList)
	assert.Equal("", node)
}

func TestClusterDetect_selectTopNode_0000x(t *testing.T) {
	assert := assert.New(t)
	nodeList := []string{
		"info_0000000046",
		"info_0000000032",
		"info_0000000058",
		"info_0000000061",
		"info_0000000008",
	}
	node := selectTopNode(nodeList)
	assert.Equal("info_0000000008", node)
}

func TestClusterDetect_selectTopNode_mixedEntries(t *testing.T) {
	assert := assert.New(t)
	nodeList := []string{
		"info_0000000046",
		"info_0000000032",
		"foo_lskdjfglsdkfsdfgdfg",
		"info_0000000061",
		"log_replicas_fdgwsdfgsdf",
		"bar",
	}
	node := selectTopNode(nodeList)
	assert.Equal("info_0000000032", node)
}
