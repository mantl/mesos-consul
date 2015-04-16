package mesos

import (
	"log"

	"github.com/CiscoCloud/mesos-consul/registry"
)

// Service map to keep track of services that we have registered
//
var servicesRegistered = make(map[string]bool)

func (m *Mesos) register(r registry.RegistryAdapter, s *registry.Service) {
	servicesRegistered[s.ID] = true

	log.Print("Registering ", s.ID)

	r.Register(s)
}

func (m *Mesos) deregister(r registry.RegistryAdapter) {
	for s, b := range servicesRegistered {
		if !b {
			log.Print("Deregistering ", s)
			r.Deregister(&registry.Service{
				ID:		s,
				})

			delete(servicesRegistered, s)
		} else {
			servicesRegistered[s] = false
		}
	}
}
