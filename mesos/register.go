package mesos

import (
	"fmt"
	"log"

	"github.com/CiscoCloud/mesos-consul/registry"

	"github.com/mesosphere/mesos-dns/records/state"
)

// Query the consul agent on the Mesos Master
// to initialize the cache.
//
// All services created by mesos-consul are prefixed
// with `mesos-consul:`
//
func (m *Mesos) LoadCache() error {
	log.Print("[DEBUG] Populating cache from Consul")

	mh := m.getLeader()

	return m.Registry.CacheLoad(mh.Ip)
}

func (m *Mesos) RegisterHosts(s state.State) {
	log.Print("[INFO] Running RegisterHosts")

	m.Agents = make(map[string]string)

	// Register slaves
	for _, f := range s.Slaves {
		agent := toIP(f.PID.Host)
		port := toPort(f.PID.Port)

		m.Agents[f.ID] = agent

		m.registerHost(&registry.Service{
			ID:      fmt.Sprintf("mesos-consul:mesos:%s:%s", f.ID, f.Hostname),
			Name:    "mesos",
			Port:    port,
			Address: agent,
			Tags:    []string{"follower"},
			Check: &registry.Check{
				HTTP:     fmt.Sprintf("http://%s:%d/slave(1)/health", agent, port),
				Interval: "10s",
			},
		})
	}

	// Register masters
	mas := m.getMasters()
	for _, ma := range mas {
		var tags []string

		if ma.IsLeader {
			tags = []string{"leader", "master"}
		} else {
			tags = []string{"master"}
		}
		s := &registry.Service{
			ID:      fmt.Sprintf("mesos-consul:mesos:%s:%s", ma.Ip, ma.PortString),
			Name:    "mesos",
			Port:    ma.Port,
			Address: ma.Ip,
			Tags:    tags,
			Check: &registry.Check{
				HTTP:     fmt.Sprintf("http://%s:%d/master/health", ma.Ip, ma.Port),
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

var ipSources = []string{"docker", "mesos", "host"}

func (m *Mesos) registerTask(t *state.Task, agent string) {
	tname := cleanName(t.Name)

	address := t.IP("docker", "mesos", "host")

	if t.Resources.PortRanges != "" {
		for _, port := range t.Resources.Ports() {
			m.Registry.Register(&registry.Service{
				ID:      fmt.Sprintf("mesos-consul:%s:%s:%s", agent, tname, port),
				Name:    tname,
				Port:    toPort(port),
				Address: address,
				Check: GetCheck(t, &CheckVar{
					Host: toIP(address),
					Port: fmt.Sprintf("%d", port),
				}),
				Agent: toIP(agent),
			})
		}
	} else {
		m.Registry.Register(&registry.Service{
			ID:      fmt.Sprintf("mesos-consul:%s-%s", agent, tname),
			Name:    tname,
			Address: address,
			Check: GetCheck(t, &CheckVar{
				Host: toIP(address),
			}),
			Agent: toIP(agent),
		})
	}
}
