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

func ZKdetect(zk string) *MesosLeader {
	if zk == "" {
		return nil
	}

	l := new(MesosLeader)

	dr, err := zkDetect(zk, l)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Print("Waiting for initial information from Zookeeper")
	select {
	case <-dr:
		log.Print("Done waiting for initial information from zookeeper")
	case <-time.After(2 * time.Minute):
		log.Fatal("Timed out waiting for initial ZK detection")
	}

	return l
}

func zkDetect(zk string, l *MesosLeader) (<-chan struct{}, error) {
	log.Print("Starting master detector for ZK ", zk)
	md, err := detector.New(zk)
	if err != nil {
		return nil, fmt.Errorf("failed to create master detector: %v", err)
	}

	var startedOnce sync.Once
	started := make(chan struct{})
	if err := md.Detect(detector.OnMasterChanged(func(info *mesosproto.MasterInfo) {
		l.leaderLock.Lock()
		defer l.leaderLock.Unlock()
		if (info == nil) {
			l.host = ""
		} else if host := info.GetHostname(); host != "" {
			ip, err := net.LookupIP(host)
			if err != nil {
				l.host = host
			} else {
				l.host = ip[0].String()
			}
		} else {
			octets := make([]byte, 4, 4)
			binary.BigEndian.PutUint32(octets, info.GetIp())
			ipv4 := net.IP(octets)
			l.host = ipv4.String()
		}
		if len(l.host) > 0 {
			l.port = fmt.Sprint(info.GetPort())
		}
		startedOnce.Do(func() { close(started) })
	})); err != nil {
		return nil, fmt.Errorf("failed to initalize master detector: %v", err)
	}
	return started, nil
}
