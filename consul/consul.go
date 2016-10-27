package consul

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

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

	config.HttpClient.Timeout = time.Duration(c.config.timeout) * time.Second
	log.Debugf("consul timeout: %d", config.HttpClient.Timeout)

	if c.config.token != "" {
		log.Debugf("setting token to %s", c.config.token)
		config.Token = c.config.token
	}

	if c.config.sslEnabled {
		log.Debugf("enabling SSL")
		config.Scheme = "https"

		tlsClientConfig := &tls.Config{}

		if c.config.sslCert != "" {
			log.Debug("Enabling SSL cert")

			//rawCert, _ := ioutil.ReadFile(c.config.sslCert)
			//prvKey := x509.ParsePKCS1PrivateKey(rawCert)
			cert, err := tls.LoadX509KeyPair(c.config.sslCert, c.config.sslCert)
			if err != nil {
				log.Fatal(err)
			}

			tlsClientConfig.Certificates = []tls.Certificate{cert}
		}

		if c.config.sslCaCert != "" {
			log.Debug("Enabling SSL CA certs")

			caCert, err := ioutil.ReadFile(c.config.sslCaCert)
			if err != nil {
				log.Fatal(err)
			}

			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)

			tlsClientConfig.RootCAs = caCertPool
		}

		if !c.config.sslVerify {
			tlsClientConfig.InsecureSkipVerify = true
		}

		config.HttpClient.Transport = &http.Transport{
			TLSClientConfig: tlsClientConfig,
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

func (c *Consul) Register(service *registry.Service) {
	if _, ok := serviceCache[service.ID]; ok {
		log.Debugf("Service found. Not registering: %s", service.ID)
		c.CacheMark(service.ID)
		return
	}

	if _, ok := c.agents[service.Agent]; !ok {
		// Agent connection not saved. Connect.
		c.agents[service.Agent] = c.newAgent(service.Agent)
	}

	log.Info("Registering ", service.ID)

	s := &consulapi.AgentServiceRegistration{
		ID:      service.ID,
		Name:    service.Name,
		Port:    service.Port,
		Address: service.Address,
		Check: &consulapi.AgentServiceCheck{
			TTL:      service.Check.TTL,
			Script:   service.Check.Script,
			HTTP:     service.Check.HTTP,
			Interval: service.Check.Interval,
		},
	}

	if len(service.Tags) > 0 {
		s.Tags = service.Tags
	}

	err := c.agents[service.Agent].Agent().ServiceRegister(s)
	if err != nil {
		log.Warnf("Unable to register %s: %s", s.ID, err.Error())
		return
	}

	serviceCache[s.ID] = newCacheEntry(s, service.Agent)
	c.CacheMark(s.ID)
}

// Deregister()
//   Deregister services that no longer exist
//
func (c *Consul) Deregister() {
	for s, b := range serviceCache {
		if c.CacheIsValid(s) {
			c.CacheProcessDeregister(s)
		} else {
			log.Infof("Deregistering %s", s)
			err := c.deregister(b.agent, b.service)
			if err != nil {
				log.Info("Deregistration error ", err)
			} else {
				delete(serviceCache, s)
			}
		}
	}
}

func (c *Consul) deregister(agent string, service *consulapi.AgentServiceRegistration) error {
	if _, ok := c.agents[agent]; !ok {
		// Agent connection not saved. Connect.
		c.agents[agent] = c.newAgent(agent)
	}

	return c.agents[agent].Agent().ServiceDeregister(service.ID)
}
