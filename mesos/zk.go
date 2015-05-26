package mesos

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/mesos/mesos-go/mesosproto"

	zoo "github.com/CiscoCloud/mesos-consul/mesos/zkdetect"
)

func (m *Mesos) zkDetector(zkURI string) {
	if (zkURI == "") {
		log.Fatal("[ERROR] Zookeeper address not provided")
	}

	dr, err := m.leaderDetect(zkURI)
	if err != nil {
		log.Fatal("[ERROR] ", err.Error())
	}

	log.Print("[INFO] Waiting for initial leader information from Zookeeper")
	select {
	case <-dr:
		log.Print("[INFO] Done waiting for initial leader information from Zookeeper")
	case <-time.After(2 * time.Minute):
		log.Fatal("[ERROR] Timed out waiting for initial ZK detection")
	}
}

func (m *Mesos) leaderDetect(zkURI string) (<-chan struct{}, error) {
	log.Print("[INFO] Starting leader detector for ZK ", zkURI)
	md, err := zoo.NewClusterDetector(zkURI)
	if err != nil {
		return nil, fmt.Errorf("failed to create master detector: %v", err)
	}

	var startedOnce sync.Once
	started := make(chan struct{})
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
				for i := range(ip) {
					four = i.To4()
					if four != nil {
						ipstring = i.String()
						break
					}
				}
				// If control reaches here there are no IPv4 addresses
				// returned by net.LookupIP. Use the hostname as ipstring
				//
				ipstring = host
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
