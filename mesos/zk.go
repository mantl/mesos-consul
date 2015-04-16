package mesos

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/mesos/mesos-go/detector"
	_ "github.com/mesos/mesos-go/detector/zoo"
	"github.com/mesos/mesos-go/mesosproto"
)

func (m *Mesos) zkDetector(zkURI string) {
	if (zkURI == "") {
		log.Fatal("Zookeeper address not provided")
	}

	dr, err := m.leaderDetect(zkURI)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Print("Waiting for initial leader information from Zookeeper")
	select {
	case <-dr:
		log.Print("Done waiting for initial leader information from Zookeeper")
	case <-time.After(2 * time.Minute):
		log.Fatal("Timed out waiting for initial ZK detection")
	}
}

func (m *Mesos) leaderDetect(zkURI string) (<-chan struct{}, error) {
	log.Print("Starting leader detector for ZK ", zkURI)
	md, err := detector.New(zkURI)
	if err != nil {
		return nil, fmt.Errorf("failed to create master detector: %v", err)
	}

	var startedOnce sync.Once
	started := make(chan struct{})
	if err := md.Detect(detector.OnMasterChanged(func(info *mesosproto.MasterInfo) {
		m.Lock.Lock()
		defer m.Lock.Unlock()
		if (info == nil) {
			m.Leader.host = ""
		} else if host := info.GetHostname(); host != "" {
			ip, err := net.LookupIP(host)
			if err != nil {
				m.Leader.host = host
			} else {
				m.Leader.host = ip[0].String()
			}
		} else {
			octets := make([]byte, 4, 4)
			binary.BigEndian.PutUint32(octets, info.GetIp())
			ipv4 := net.IP(octets)
			m.Leader.host = ipv4.String()
		}
		if len(m.Leader.host) > 0 {
			m.Leader.port = fmt.Sprint(info.GetPort())
		}
		startedOnce.Do(func() { close(started) })
	})); err != nil {
		return nil, fmt.Errorf("failed to initalize master detector: %v", err)
	}
	return started, nil
}

func (m *Mesos) getLeader() (string, string) {
	m.Lock.Lock()
	defer m.Lock.Unlock()
	return m.Leader.host, m.Leader.port
}
