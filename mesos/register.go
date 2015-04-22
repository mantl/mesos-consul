package mesos

import (
	"fmt"
	"log"

	"github.com/CiscoCloud/mesos-consul/registry"
)

type registryState struct {
	services	map[string]bool
	followers	map[string]bool
	masters		map[string]MesosHost
}


var currentState = registryState{
	services:	make(map[string]bool),
	followers:	make(map[string]bool),
	masters:	make(map[string]MesosHost),
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

			c := new(registry.ServiceCheck)
			c.HTTP = fmt.Sprintf("%s:%d/slave(1)/health", s.IP, s.Port)
			c.Interval = "10s"
			s.Check = c

			log.Print("Registering: ", id)

			err := m.registry.Register(s)
			if err != nil {
				log.Print(err)
			}
		}
	}

	// Register masters
	mas := m.getMasters()
	for _, ma := range mas {
		id := fmt.Sprintf("%s:%s", ma.host, ma.port)
		if master, ok := currentState.masters[id]; ok {
			// Master has been found. Only update if the leader status has changed
			if ma.isLeader ==  master.isLeader {
				master.isRegistered = true
				log.Printf("Master state unchanged, not registering: '%s'", id)
				continue
			}
		}

		ma.isRegistered = true
		currentState.masters[id] = ma

		s := new(registry.Service)
		s.ID	= fmt.Sprintf("mesos:%s:%s", ma.host, ma.port)
		s.Name	= "mesos"
		s.Port	= toPort(ma.port)
		s.IP	= toIP(ma.host)
		if ma.isLeader {
			s.Tags	= []string{
				"master",
				}
		} else {
			s.Tags	= []string{
				"master",
				"leader",
				}
		}

		c := new(registry.ServiceCheck)
		c.HTTP = fmt.Sprintf("%s:%d/master/health", s.IP, s.Port)
		c.Interval = "10s"
		s.Check = c

		log.Print("Registering: ", s.ID)
		err := m.registry.Register(s)
		if err != nil {
			log.Print(err)
		}
	}
}

func (m *Mesos) register(s *registry.Service) {
	if _, ok := currentState.services[s.ID]; ok {
		log.Printf("Service found. Not registering: %s", s.ID)
		currentState.services[s.ID] = true
		return
	}

	log.Print("Registering ", s.ID)

	currentState.services[s.ID] = true
	err := m.registry.Register(s)
	if err != nil {
		log.Print(err)
	}
}

// deregister items that have gone away
//
func (m *Mesos) deregister() {
	// Services
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

	// Followers
	for id, isRegistered := range currentState.followers {
		if !isRegistered {
			log.Print("Deregistering ", id)
			m.registry.Deregister(&registry.Service{
				ID:		id,
				})
			delete(currentState.followers, id)
		} else {
			currentState.followers[id] = false
		}
	}

	// Masters
	for id, master := range currentState.masters {
		if !master.isRegistered {
			log.Print("Deregistering ", id)
			m.registry.Deregister(&registry.Service{
				ID:		id,
				})
			delete(currentState.masters, id)
		} else {
			master.isRegistered = false
		}
	}
}
