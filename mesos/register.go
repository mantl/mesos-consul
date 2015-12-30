package mesos

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/CiscoCloud/mesos-consul/registry"
	"github.com/CiscoCloud/mesos-consul/state"

	log "github.com/sirupsen/logrus"
)

// Query the consul agent on the Mesos Master
// to initialize the cache.
//
// All services created by mesos-consul are prefixed
// with `mesos-consul:`
//
func (m *Mesos) LoadCache() error {
	log.Debug("Populating cache from Consul")

	mh := m.getLeader()

	return m.Registry.CacheLoad(mh.Ip)
}

func (m *Mesos) RegisterHosts(s state.State) {
	log.Debug("Running RegisterHosts")

	m.Agents = make(map[string]string)

	// Register slaves
	for _, f := range s.Slaves {
		agent := toIP(f.PID.Host)
		port := toPort(f.PID.Port)

		m.Agents[f.ID] = agent

		m.registerHost(&registry.Service{
			ID:      fmt.Sprintf("mesos-consul:%s:%s:%s", m.ServiceName, f.ID, f.Hostname),
			Name:    m.ServiceName,
			Port:    port,
			Address: agent,
			Agent:   agent,
			Tags:    m.agentTags("agent", "follower"),
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
			tags = m.agentTags("leader", "master")
		} else {
			tags = m.agentTags("master")
		}
		s := &registry.Service{
			ID:      fmt.Sprintf("mesos-consul:%s:%s:%s", m.ServiceName, ma.Ip, ma.PortString),
			Name:    m.ServiceName,
			Port:    ma.Port,
			Address: ma.Ip,
			Agent:   ma.Ip,
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
		log.Infof("Host found. Comparing tags: (%v, %v)", h.Tags, s.Tags)

		if sliceEq(s.Tags, h.Tags) {
			m.Registry.CacheMark(s.ID)

			// Tags are the same. Return
			return
		}

		log.Info("Tags changed. Re-registering")

		// Delete cache entry. It will be re-created below
		m.Registry.CacheDelete(s.ID)
	}

	m.Registry.Register(s)
}

func (m *Mesos) registerTask(t *state.Task, agent string) {
	var tags []string

	tname := cleanName(t.Name)
	if m.whitelistRegex != nil {
		if !m.whitelistRegex.MatchString(tname) {
			log.WithField("task", tname).Debug("Task not on whitelist")
			// No match
			return
		}
	}

	address := t.IP(m.IpOrder...)

	l := t.Label("tags")
	if l != "" {
		tags = strings.Split(t.Label("tags"), ",")
	} else {
		tags = []string{}
	}

	for key := range t.DiscoveryInfo.Ports.DiscoveryPorts {
		// We append -portN to ports after the first.
		// This is done to preserve compatibility with
		// existing implementations which may rely on the
		// old unprefixed name.
		svcName := tname
		if key > 0 {
			svcName = fmt.Sprintf("%s-port%d", svcName, key+1)
		}
		discoveryPort := state.DiscoveryPort(t.DiscoveryInfo.Ports.DiscoveryPorts[key])
		serviceName := discoveryPort.Name
		servicePort := strconv.Itoa(discoveryPort.Number)
		log.Debugf("%+v framework has %+v as a name for %+v port",
			t.Name,
			discoveryPort.Name,
			discoveryPort.Number)
		if discoveryPort.Name != "" {
			m.Registry.Register(&registry.Service{
				ID:      fmt.Sprintf("mesos-consul:%s:%s:%d", agent, svcName, discoveryPort.Number),
				Name:    svcName,
				Port:    toPort(servicePort),
				Address: address,
				Tags:    []string{serviceName},
				Check: GetCheck(t, &CheckVar{
					Host: toIP(address),
					Port: servicePort,
				}),
				Agent: toIP(agent),
			})
		}
	}

	if t.Resources.PortRanges != "" {
		for key, port := range t.Resources.Ports() {
			// We append -portN to ports after the first.
			// This is done to preserve compatibility with
			// existing implementations which may rely on the
			// old unprefixed name.
			svcName := tname
			if key > 0 {
				svcName = fmt.Sprintf("%s-port%d", svcName, key+1)
			}
			m.Registry.Register(&registry.Service{
				ID:      fmt.Sprintf("mesos-consul:%s:%s:%s", agent, svcName, port),
				Name:    svcName,
				Port:    toPort(port),
				Address: address,
				Tags:    tags,
				Check: GetCheck(t, &CheckVar{
					Host: toIP(address),
					Port: port,
				}),
				Agent: toIP(agent),
			})
		}
	} else {
		m.Registry.Register(&registry.Service{
			ID:      fmt.Sprintf("mesos-consul:%s-%s", agent, tname),
			Name:    tname,
			Address: address,
			Tags:    tags,
			Check: GetCheck(t, &CheckVar{
				Host: toIP(address),
			}),
			Agent: toIP(agent),
		})
	}
}

func (m *Mesos) agentTags(ts ...string) []string {
	if len(m.ServiceTags) == 0 {
		return ts
	}

	rval := []string{}

	for _, tag := range m.ServiceTags {
		for _, t := range ts {
			rval = append(rval, fmt.Sprintf("%s.%s", t, tag))
		}
	}

	return rval
}
