package consul

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"

	"github.com/CiscoCloud/mesos-consul/registry"

	consulapi "github.com/hashicorp/consul/api"
)

type Consul struct {
	agents		map[string]*consulapi.Client
	config		consulConfig
}

//
func New() *Consul {
	return &Consul{
		agents:		make(map[string]*consulapi.Client),
		config:		config,
	}
}

// client()
//   Return a consul client at the specified address
func (c *Consul) client(address string) *consulapi.Client {
	if address == "" {
		log.Print("[WARN] No address to Consul.Agent")
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
		log.Printf("[WARN] No address to Consul.NewAgent")
		return nil
	}

	config := consulapi.DefaultConfig()

	config.Address = fmt.Sprintf("%s:%s", address, c.config.port)

	if c.config.token != "" {
		log.Printf("[DEBUG] setting token to %s", c.config.token)
		config.Token = c.config.token
	}

	if c.config.sslEnabled {
		log.Printf("[DEBUG] enabling SSL")
		config.Scheme = "https"
	}

	if !c.config.sslVerify {
		log.Printf("[DEBUG] disabled SSL verification")
		config.HttpClient.Transport = &http.Transport {
			TLSClientConfig: &tls.Config {
				InsecureSkipVerify: true,
			},
		}
	}

	if c.config.auth.Enabled {
		log.Printf("[DEBUG] setting basic auth")
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
	if _, ok := serviceCache[service.ID]; ok {
		log.Printf("[DEBUG] Service found. Not registering: %s", service.ID)
		serviceCache[service.ID].isRegistered = true
		return nil
	}

	if _, ok := c.agents[service.Address]; !ok {
		// Agent connection not saved. Connect.
		c.agents[service.Address] = c.newAgent(service.Address)
	}

	log.Print("[INFO] Registering ", service.ID)

	s := &consulapi.AgentServiceRegistration{
		ID:		service.ID,
		Name:		service.Name,
		Port:		service.Port,
		Address:	service.Address,
		Tags:		service.Tags,
		Check:		&consulapi.AgentServiceCheck{
			TTL:		service.Check.TTL,
			Script:		service.Check.Script,
			HTTP:		service.Check.HTTP,
			Interval:	service.Check.Interval,
		},
	}

	serviceCache[s.ID] = &cacheEntry{
		service:	s,
		isRegistered:	true,
	}

	return c.agents[service.Address].Agent().ServiceRegister(s)
}

// Deregister()
//   Deregister services that no longer exist
//
func (c *Consul) Deregister() error {
	for s, b := range serviceCache {
		if !b.isRegistered {
			log.Printf("[INFO] Deregistering %s", s)
			err := c.deregister(b.service)
			if err != nil {
				return err
			}
			delete(serviceCache, s)
		} else {
			serviceCache[s].isRegistered = true
		}
	}

	return nil
}

func (c *Consul) deregister(service *consulapi.AgentServiceRegistration) error {
	if _, ok := c.agents[service.Address]; !ok {
		log.Print("[WARN] Deregistering a service without an agent connection?!")

		// Agent connection not saved. Connect.
		c.agents[service.Address] = c.newAgent(service.Address)
	}

	return c.agents[service.Address].Agent().ServiceDeregister(service.ID)
}
