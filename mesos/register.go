package mesos

import (
	"fmt"
	"log"

	consulapi "github.com/hashicorp/consul/api"
)

type cacheEntry struct {
	service		*consulapi.AgentServiceRegistration
	isRegistered	bool
}

var cache = make(map[string]*cacheEntry)

func (m *Mesos) RegisterHosts(sj StateJSON) {
	log.Print("[INFO] Running RegisterHosts")

	// Register followers
	for _, f := range sj.Followers {
		h, p := parsePID(f.Pid)
		host := toIP(h)
		port := toPort(p)

		m.register(&consulapi.AgentServiceRegistration{
			ID:		fmt.Sprintf("%s:%s", f.Id, f.Hostname),
			Name:		"mesos",
			Port:		port,
			Address:	host,
			Tags:		[]string{ "follower" },
			Check:		&consulapi.AgentServiceCheck{
				HTTP:		fmt.Sprintf("http://%s:%d/slave(1)/health", host, port),
				Interval:	"10s",
			},
		})
	}

	// Register masters
	mas := m.getMasters()
	for _, ma := range mas {
		var tags []string

		if ma.isLeader {
			tags = []string{ "leader", "master" }
		} else {
			tags = []string{ "master" }
		}
		host := toIP(ma.host)
		port := toPort(ma.port)
		s := &consulapi.AgentServiceRegistration{
			ID:		fmt.Sprintf("mesos:%s:%s", ma.host, ma.port),
			Name:		"mesos",
			Port:		port,
			Address:	host,
			Tags:		tags,
			Check:		&consulapi.AgentServiceCheck{
				HTTP:		fmt.Sprintf("http://%s:%d/master/health", host, port),
				Interval:	"10s",
			},
		}

		err := m.Consul.Register(s)
		if err != nil {
			log.Print("[ERROR] ", err)
		}
	}
}

func (m *Mesos) register(s *consulapi.AgentServiceRegistration) {
	if _, ok := cache[s.ID]; ok {
		log.Printf("[INFO] Service found. Not registering: %s", s.ID)
		cache[s.ID].isRegistered = true
		return
	}

	log.Print("[INFO] Registering ", s.ID)

	cache[s.ID] = &cacheEntry{
		service:		s,
		isRegistered:		true,
	}

	err := m.Consul.Register(s)
	if err != nil {
		log.Print("[ERROR] ", err)
	}
}

// deregister items that have gone away
//
func (m *Mesos) deregister() {
	for s, b := range cache {
		if !b.isRegistered {
			log.Print("[INFO] Deregistering ", s)
			m.Consul.Deregister(b.service)

			delete(cache, s)
		} else {
			cache[s].isRegistered = false
		}
	}
}
