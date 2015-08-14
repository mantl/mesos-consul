package mesos

import (
	"fmt"
	"log"

	"github.com/CiscoCloud/mesos-consul/registry"
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

	return m.Registry.CacheLoad(host)
}

func (m *Mesos) RegisterHosts(sj StateJSON) {
	log.Print("[INFO] Running RegisterHosts")

	// Register followers
	for _, f := range sj.Followers {
		h, p := parsePID(f.Pid)
		host := toIP(h)
		port := toPort(p)

		m.registerHost(&registry.Service{
			ID:      fmt.Sprintf("mesos-consul:mesos:%s:%s", f.Id, f.Hostname),
			Name:    "mesos",
			Port:    port,
			Address: host,
			Tags:    []string{"follower"},
			Check: &registry.Check{
				HTTP:     fmt.Sprintf("http://%s:%d/slave(1)/health", host, port),
				Interval: "10s",
			},
		})
	}

	// Register masters
	mas := m.getMasters()
	for _, ma := range mas {
		var tags []string

		if ma.isLeader {
			tags = []string{"leader", "master"}
		} else {
			tags = []string{"master"}
		}
		host := toIP(ma.host)
		port := toPort(ma.port)
		s := &registry.Service{
			ID:      fmt.Sprintf("mesos-consul:mesos:%s:%s", ma.host, ma.port),
			Name:    "mesos",
			Port:    port,
			Address: host,
			Tags:    tags,
			Check: &registry.Check{
				HTTP:     fmt.Sprintf("http://%s:%d/master/health", host, port),
				Interval: "10s",
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

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func (m *Mesos) registerHost(s *registry.Service) {
	h := m.Registry.CacheLookup(s.ID)
	if h != nil {
		log.Printf("[INFO] Host found. Comparing tags: (%v, %v)", h.Tags, s.Tags)

		if sliceEq(s.Tags, h.Tags) {
			m.Registry.CacheMark(s.ID)

			// Tags are the same. Return
			return
		}

		log.Println("[INFO] Tags changed. Re-registering")

		// Delete cache entry. It will be re-created below
		m.Registry.CacheDelete(s.ID)
	}

	err := m.Registry.Register(s)
	if err != nil {
		log.Print("[ERROR] ", err)
	}
}

func (m *Mesos) registerTask(t *Task, host string) {
	tname := cleanName(t.Name)

	if t.Resources.Ports != "" {
		for _, port := range yankPorts(t.Resources.Ports) {
			m.Registry.Register(&registry.Service{
				ID:      fmt.Sprintf("mesos-consul:%s:%s:%d", host, tname, port),
				Name:    tname,
				Port:    port,
				Address: toIP(host),
				Check:   t.GetCheck(),
			})
		}
	} else {
		m.Registry.Register(&registry.Service{
			ID:      fmt.Sprintf("mesos-consul:%s-%s", host, tname),
			Name:    tname,
			Address: toIP(host),
			Check:   t.GetCheck(),
		})
	}
}
