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

		m.Masters = new([]MesosHost)

		// Handle list of masters
		for _, ma := range *info.Masters {
			mh := m.hostFromMasterInfo(ma)

			*m.Masters = append(*m.Masters, mh)
		}

		// Handle leader
		ma := m.hostFromMasterInfo(info.Leader)
		if len(ma.host) > 0 {
			ma.isLeader = true
		}

		*m.Masters = append(*m.Masters, ma)

		startedOnce.Do(func() { close(started) })
	})); err != nil {
		return nil, fmt.Errorf("failed to initalize master detector: %v", err)
	}
	return started, nil
}

// Get the leader out of the list of masters
//
func (m *Mesos) getLeader() (string, string) {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	for _, ms := range *m.Masters {
		if ms.isLeader {
			return ms.host, ms.port
		}
	}

	return "", ""
}

func (m *Mesos) getMasters() []MesosHost {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	ms := make([]MesosHost, len(*m.Masters))
	for i, msp := range *m.Masters {
		mh := MesosHost{
			host:		msp.host,
			port:		msp.port,
			isLeader:	msp.isLeader,
		}
			
		ms[i] = mh
	}
	return ms
}

func (m *Mesos) hostFromMasterInfo(mi *mesosproto.MasterInfo) MesosHost {
	var ipstring = ""
	var port = ""

	if mi != nil {
		if host := mi.GetHostname(); host != "" {
			ip, err := net.LookupIP(host)
			if err != nil {
				ipstring = host
			} else {
				ipstring = ip[0].String()
			}
		} else {
			octets := make([]byte, 4, 4)
			binary.BigEndian.PutUint32(octets, mi.GetIp())
			ipv4 := net.IP(octets)
			ipstring = ipv4.String()
		}
	}

	if len(ipstring) > 0 {
		port = fmt.Sprint(mi.GetPort())
	}
		

	return MesosHost{
		host:		ipstring,
		port:		port,
		isLeader:	false,
	}
}
