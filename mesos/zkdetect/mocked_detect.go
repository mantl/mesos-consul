package zkdetect

import (
	"errors"
	"fmt"
	"log"
	"net/url"

	"github.com/gogo/protobuf/proto"
	util "github.com/mesos/mesos-go/mesosutil"
	"github.com/samuel/go-zookeeper/zk"
)

type MockClusterDetector struct {
	*ClusterDetector
	zkPath string
	conCh  chan zk.Event
	sesCh  chan zk.Event
}

func NewMockClusterDetector(zkurls string) (*MockClusterDetector, error) {
	log.Print("[INFO] Creating mock zk master detector")
	md, err := NewClusterDetector(zkurls)
	if err != nil {
		return nil, err
	}

	u, _ := url.Parse(zkurls)
	m := &MockClusterDetector{
		ClusterDetector: md,
		zkPath:         u.Path,
		conCh:          make(chan zk.Event, 5),
		sesCh:          make(chan zk.Event, 5),
	}

	path := m.zkPath
	connector := NewMockConnector()
	connector.On("Children", path).Return([]string{"info_0", "info_5", "info_10"}, &zk.Stat{}, nil)
	connector.On("Get", fmt.Sprintf("%s/info_0", path)).Return(m.makeClusterInfo(), &zk.Stat{}, nil)
	connector.On("Close").Return(nil)
	connector.On("ChildrenW", m.zkPath).Return([]string{m.zkPath}, &zk.Stat{}, (<-chan zk.Event)(m.sesCh), nil)

	first := true
	m.client.setFactory(asFactory(func() (Connector, <-chan zk.Event, error) {
		if !first {
			return nil, nil, errors.New("only 1 connector allowed")
		} else {
			first = false
		}
		return connector, m.conCh, nil
	}))

	return m, nil
}

func (m *MockClusterDetector) Start() {
	m.client.connect()
}

func (m *MockClusterDetector) ScheduleConnEvent(s zk.State) {
	log.Printf("[INFO] Scheduling zk connection event with state: %v\n", s)
	go func() {
		m.conCh <- zk.Event{
			State: s,
			Path:  m.zkPath,
		}
	}()
}

func (m *MockClusterDetector) ScheduleSessEvent(t zk.EventType) {
	log.Printf("[INFO] Scheduling zk session event with state: %v\n", t)
	go func() {
		m.sesCh <- zk.Event{
			Type: t,
			Path: m.zkPath,
		}
	}()
}

func (m *MockClusterDetector) makeClusterInfo() []byte {
	miPb := util.NewMasterInfo("master", 123456789, 400)
	miPb.Pid = proto.String("master@127.0.0.1:5050")
	data, err := proto.Marshal(miPb)
	if err != nil {
		panic(err)
	}
	return data
}
