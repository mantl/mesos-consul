package mesos

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	//"github.com/mesos/mesos-go/detector"
	//_ "github.com/mesos/mesos-go/detector/zoo"
	"github.com/mesos/mesos-go/mesosproto"

	zoo "github.com/CiscoCloud/mesos-consul/mesos/zkdetect"
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
	//md, err := detector.New(zkURI)
	md, err := zoo.NewClusterDetector(zkURI)
	if err != nil {
		return nil, fmt.Errorf("failed to create master detector: %v", err)
	}

	var startedOnce sync.Once
	started := make(chan struct{})
//	if err := md.Detect(detector.OnMasterChanged(func(info *mesosproto.MasterInfo) {
	if err := md.Detect(zoo.OnClusterChanged(func(info *zoo.ClusterInfo) {
		m.Lock.Lock()
		defer m.Lock.Unlock()

		m.Masters = new([]*MesosHost)

		// Handle list of masters
		for _, ma := range *info.Masters {
			mh := new(MesosHost)

			mh.host = m.getMasterIp(ma)
			if len(mh.host) > 0 {
				mh.port = fmt.Sprint(ma.GetPort())
			}

			*m.Masters = append(*m.Masters, mh)
		}

		m.Leader = new(MesosHost)
		// Handle leader
		if (info.Leader == nil) {
			m.Leader.host = ""
		} else {
			m.Leader.host = m.getMasterIp(info.Leader)
			if len(m.Leader.host) > 0 {
				m.Leader.port = fmt.Sprint(info.Leader.GetPort())
			}
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

func (m *Mesos) getMasters() []*MesosHost {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	ms := make([]*MesosHost, len(*m.Masters))
	for _,msp := range ms {
		mh := new(MesosHost)
		mh.host = msp.host
		mh.port = msp.port
		ms = append(ms, mh)
	}
	return ms
}

func (ms *Mesos) getMasterIp(m *mesosproto.MasterInfo) string {
	if m == nil {
		return ""
	}

	if host := m.GetHostname(); host != "" {
		ip, err := net.LookupIP(host)
		if err != nil {
			return host
		} else {
			return ip[0].String()
		}
	} else {
		octets := make([]byte, 4, 4)
		binary.BigEndian.PutUint32(octets, m.GetIp())
		ipv4 := net.IP(octets)
		return ipv4.String()
	}

	return ""
}
