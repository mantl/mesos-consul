package mesos

import (
	"fmt"
	"log"
	"strings"

	consulapi "github.com/hashicorp/consul/api"
)

// Query the consul agent on the Mesos Master
// to initialize the cache.
//
// All services created by mesos-consul are prefixed
// with `mesos-consul:`
//
func (m *Mesos) LoadCache() error {
	log.Print("[DEBUG] Populating cache from Consul")

	host, _ := m.getLeader()
	
	client := m.Consul.Client(host).Catalog()

	serviceList, _, err := client.Services(nil)
	if err != nil {
		return err
	}

	for service, _ := range serviceList {
		catalogServices, _, err := client.Service(service, "", nil)
		if err != nil {
			return err
		}

		for _, s := range catalogServices {
			if strings.HasPrefix(s.ServiceID, "mesos-consul:")  {
				log.Printf("[DEBUG] Found '%s' with ID '%s'", s.ServiceName, s.ServiceID)
				m.ServiceCache[s.ServiceID] = &CacheEntry{
					service:	&consulapi.AgentServiceRegistration{
							ID:		s.ServiceID,
							Name:		s.ServiceName,
							Port:		s.ServicePort,
							Address:	s.ServiceAddress,
							Tags:		s.ServiceTags,
							},
					isRegistered:	false,
				}
			}
		}
	}

	return nil
}

func (m *Mesos) RegisterHosts(sj StateJSON) {
	log.Print("[INFO] Running RegisterHosts")

	// Register followers
	for _, f := range sj.Followers {
		h, p := parsePID(f.Pid)
		host := toIP(h)
		port := toPort(p)

		m.registerHost(&consulapi.AgentServiceRegistration{
			ID:		fmt.Sprintf("mesos-consul:mesos:%s:%s", f.Id, f.Hostname),
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
			ID:		fmt.Sprintf("mesos-consul:mesos:%s:%s", ma.host, ma.port),
			Name:		"mesos",
			Port:		port,
			Address:	host,
			Tags:		tags,
			Check:		&consulapi.AgentServiceCheck{
				HTTP:		fmt.Sprintf("http://%s:%d/master/health", host, port),
				Interval:	"10s",
			},
		}

		m.registerHost(s)
	}
}

// helper function to compare service tag slices
//
func sliceEq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a{
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func (m *Mesos) registerHost(s *consulapi.AgentServiceRegistration) {

	if _, ok := m.ServiceCache[s.ID]; ok {
		log.Printf("[INFO] Host found. Comparing tags: (%v, %v)", m.ServiceCache[s.ID].service.Tags, s.Tags)

		if sliceEq(s.Tags, m.ServiceCache[s.ID].service.Tags) {
			m.ServiceCache[s.ID].isRegistered = true

			// Tags are the same. Return
			return
		}

		log.Println("[INFO] Tags changed. Re-registering")

		// Delete cache entry. It will be re-created below
		delete(m.ServiceCache, s.ID)
	}

	m.ServiceCache[s.ID] = &CacheEntry{
		service:		s,
		isRegistered:		true,
	}


	err := m.Consul.Register(s)
	if err != nil {
		log.Print("[ERROR] ", err)
	}
}

func (m *Mesos) register(s *consulapi.AgentServiceRegistration) {
	if _, ok := m.ServiceCache[s.ID]; ok {
		log.Printf("[INFO] Service found. Not registering: %s", s.ID)
		m.ServiceCache[s.ID].isRegistered = true
		return
	}

	log.Print("[INFO] Registering ", s.ID)

	m.ServiceCache[s.ID] = &CacheEntry{
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
	for s, b := range m.ServiceCache {
		if !b.isRegistered {
			log.Print("[INFO] Deregistering ", s)
			m.Consul.Deregister(b.service)

			delete(m.ServiceCache, s)
		} else {
			m.ServiceCache[s].isRegistered = false
		}
	}
}
