package mesos

import (
	"fmt"
	"log"

	"github.com/CiscoCloud/mesos-consul/registry"
)

type registryState struct {
	services	map[string]bool
	followers	map[string]bool
}


var currentState = registryState{
	services:	make(map[string]bool),
	followers:	make(map[string]bool),
}

func (m *Mesos) RegisterHosts(sj StateJSON) {
	log.Print("Running RegisterHosts")

	// Register followers
	for _, f := range sj.Followers {
		id := fmt.Sprintf("%s:%s", f.Id, f.Hostname)
		if _, ok := currentState.followers[id]; ok {
			log.Printf("Id found: '%s': Not registering", id)
			currentState.followers[id] = true
		} else {
			currentState.followers[id] = true

			s := new(registry.Service)

			host, port := parsePID(f.Pid)

			s.ID	= id
			s.Name	= "mesos"
			s.Port	= toPort(port)
			s.IP	= toIP(host)
			s.Tags	= []string{
				"follower",
				}

			log.Print("Registering: ", id)

			m.registry.Register(s)
			if err != nil {
				log.Print(err)
			}
		}
	}

	// Register masters
	// TODO

	// Register leader
	ip, port := m.getLeader()

	s := new(registry.Service)
	s.ID	= fmt.Sprintf("mesos:%s:%s", ip, port)
	s.Name	= "mesos"
	s.Port	= toPort(port)
	s.IP	= toIP(ip)
	s.Tags	= []string{
		"leader",
		"master",
		}
		

	log.Print("Registering: ", s.ID)
	err = m.registry.Register(s)
	if err != nil {
		log.Print(err)
	}
}

func (m *Mesos) register(s *registry.Service) {
	currentState.services[s.ID] = true

	log.Print("Registering ", s.ID)

	err := m.registry.Register(s)
	if err != nil {
		log.Print(err)
	}
}

func (m *Mesos) deregister() {
	for s, b := range currentState.services {
		if !b {
			log.Print("Deregistering ", s)
			m.registry.Deregister(&registry.Service{
				ID:		s,
				})

			delete(currentState.services, s)
		} else {
			currentState.services[s] = false
		}
	}
}
