package consul

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/CiscoCloud/mesos-consul/registry"

	consulapi "github.com/hashicorp/consul/api"
	log "github.com/sirupsen/logrus"
)

type Consul struct {
	agents map[string]*consulapi.Client
	config consulConfig
}

//
func New() *Consul {
	return &Consul{
		agents: make(map[string]*consulapi.Client),
		config: config,
	}
}

// client()
//   Return a consul client at the specified address
func (c *Consul) client(address string) *consulapi.Client {
	if address == "" {
		log.Warn("No address to Consul.Agent")
		return nil
	}

	if _, ok := c.agents[address]; !ok {
		// Agent connection not saved. Connect.
		c.agents[address] = c.newAgent(address)
	}

	return c.agents[address]
}

// newAgent()
//   Connect to a new agent specified by address
//
func (c *Consul) newAgent(address string) *consulapi.Client {
	if address == "" {
		log.Warnf("No address to Consul.NewAgent")
		return nil
	}

	config := consulapi.DefaultConfig()

	config.Address = fmt.Sprintf("%s:%s", address, c.config.port)
	log.Debugf("consul address: %s", config.Address)

	if c.config.token != "" {
		log.Debugf("setting token to %s", c.config.token)
		config.Token = c.config.token
	}

	if c.config.sslEnabled {
		log.Debugf("enabling SSL")
		config.Scheme = "https"
	}

	if !c.config.sslVerify {
		log.Debugf("disabled SSL verification")
		config.HttpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}

	if c.config.auth.Enabled {
		log.Debugf("setting basic auth")
		config.HttpAuth = &consulapi.HttpBasicAuth{
			Username: c.config.auth.Username,
			Password: c.config.auth.Password,
		}
	}

	client, err := consulapi.NewClient(config)
	if err != nil {
		log.Fatal("consul: ", address)
	}
	return client
}

func (c *Consul) Register(service *registry.Service) error {
	var address string

	if _, ok := serviceCache[service.ID]; ok {
		log.Debugf("Service found. Not registering: %s", service.ID)
		serviceCache[service.ID].isRegistered = true
		return nil
	}

	if service.Agent != "" {
		address = service.Agent
	} else {
		address = service.Address
	}

	if _, ok := c.agents[address]; !ok {
		// Agent connection not saved. Connect.
		c.agents[address] = c.newAgent(address)
	}

	log.Info("Registering ", service.ID)

	s := &consulapi.AgentServiceRegistration{
		ID:      service.ID,
		Name:    service.Name,
		Port:    service.Port,
		Address: service.Address,
		Tags:    service.Tags,
		Check: &consulapi.AgentServiceCheck{
			TTL:      service.Check.TTL,
			Script:   service.Check.Script,
			HTTP:     service.Check.HTTP,
			Interval: service.Check.Interval,
		},
	}

	serviceCache[s.ID] = &cacheEntry{
		service:      s,
		isRegistered: true,
	}

	return c.agents[address].Agent().ServiceRegister(s)
}

// Deregister()
//   Deregister services that no longer exist
//
func (c *Consul) Deregister() error {
	for s, b := range serviceCache {
		if !b.isRegistered {
			log.Infof("Deregistering %s", s)
			err := c.deregister(b.service)
			if err != nil {
				return err
			}
			delete(serviceCache, s)
		} else {
			serviceCache[s].isRegistered = false
		}
	}

	return nil
}

func (c *Consul) deregister(service *consulapi.AgentServiceRegistration) error {
	if _, ok := c.agents[service.Address]; !ok {
		log.Warn("Deregistering a service without an agent connection?!")

		// Agent connection not saved. Connect.
		c.agents[service.Address] = c.newAgent(service.Address)
	}

	return c.agents[service.Address].Agent().ServiceDeregister(service.ID)
}
