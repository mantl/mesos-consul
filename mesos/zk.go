package mesos

import (
	"fmt"
	"log"

	"github.com/mesos/mesos-go/detector"
	_ "github.com/mesos/mesos-go/detector/zoo"
	proto "github.com/mesos/mesos-go/mesosproto"
)

func (m *Mesos) OnMasterChanged(leader *proto.MasterInfo) {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	m.Leader = leader
}

func (m *Mesos) UpdatedMasters(masters []*proto.MasterInfo) {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	m.Masters = masters
}

func (m *Mesos) zkDetector(zkURI string) {
	if zkURI == "" {
		log.Fatal("[ERROR] Zookeeper address not provided")
	}

	md, err := detector.New(zkURI)
	if err != nil {
		log.Fatal("[ERROR] ", err.Error())
	}

	md.Detect(m)
}

// Get the leader out of the list of masters
//
func (m *Mesos) getLeader() *MesosHost {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	return MasterInfoToMesosHost(m.Leader)
}

func (m *Mesos) getMasters() []*MesosHost {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	ms := make([]*MesosHost, len(m.Masters))
	for i, msp := range m.Masters {
		mh := MasterInfoToMesosHost(msp)
		if *m.Leader.Id == *msp.Id {
			mh.IsLeader = true
		}

		ms[i] = mh
	}
	return ms
}

func MasterInfoToMesosHost(mi *proto.MasterInfo) *MesosHost {
	if mi == nil {
		return &MesosHost{
			Host:         "",
			Ip:           "",
			Port:         0,
			PortString:   "",
			IsLeader:     false,
			IsRegistered: false,
		}
	}

	return &MesosHost{
		Host:         *mi.Address.Hostname,
		Ip:           *mi.Address.Ip,
		Port:         int(*mi.Address.Port),
		PortString:   fmt.Sprintf("%d", *mi.Address.Port),
		IsLeader:     false,
		IsRegistered: false,
	}
}
