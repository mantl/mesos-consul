package consul

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"

	"github.com/CiscoCloud/mesos-consul/config"
	consulapi "github.com/hashicorp/consul/api"
)

type Consul struct {
	agents map[string]*consulapi.Client
	config *config.Config
}

// Wraps the AgentServiceRegistration struct to allow storing agent separately
type ServiceRegistration struct {
	Agent        string
	Registration *consulapi.AgentServiceRegistration
}

// Instantiates a new ServiceRegistration struct
func NewRegistration(asr *consulapi.AgentServiceRegistration) *ServiceRegistration {
	return &ServiceRegistration{Agent: asr.Address, Registration: asr}
}

//
func NewConsul(c *config.Config) *Consul {
	return &Consul{
		agents: make(map[string]*consulapi.Client),
		config: c,
	}
}

// Client()
//   Return a consul client at the specified address
func (c *Consul) Client(address string) *consulapi.Client {
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

	config.Address = fmt.Sprintf("%s:%s", address, c.config.RegistryPort)

	if c.config.RegistryToken != "" {
		log.Printf("[DEBUG] setting token to %s", c.config.RegistryToken)
		config.Token = c.config.RegistryToken
	}

	if c.config.RegistrySSL.Enabled {
		log.Printf("[DEBUG] enabling SSL")
		config.Scheme = "https"
	}

	if !c.config.RegistrySSL.Verify {
		log.Printf("[DEBUG] disabled SSL verification")
		config.HttpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}

	if c.config.RegistryAuth.Enabled {
		log.Printf("[DEBUG] setting basic auth")
		config.HttpAuth = &consulapi.HttpBasicAuth{
			Username: c.config.RegistryAuth.Username,
			Password: c.config.RegistryAuth.Password,
		}
	}

	client, err := consulapi.NewClient(config)
	if err != nil {
		log.Fatal("consul: ", address)
	}
	return client
}

func (r *Consul) Register(sr *ServiceRegistration) error {
	if _, ok := r.agents[sr.Agent]; !ok {
		// Agent connection not saved. Connect.
		r.agents[sr.Agent] = r.newAgent(sr.Agent)
	}

	return r.agents[sr.Agent].Agent().ServiceRegister(sr.Registration)
}

func (r *Consul) Deregister(sr *ServiceRegistration) error {
	if _, ok := r.agents[sr.Agent]; !ok {
		log.Print("[WARN] Deregistering a service without an agent connection?!")

		// Agent connection not saved. Connect.
		r.agents[sr.Agent] = r.newAgent(sr.Agent)
	}

	return r.agents[sr.Agent].Agent().ServiceDeregister(sr.Registration.ID)
}
