package consul

import (
	"crypto/tls"
	"log"
	"net/http"
	"net/url"

	"github.com/CiscoCloud/mesos-consul/registry"
	"github.com/CiscoCloud/mesos-consul/config"
	consulapi "github.com/hashicorp/consul/api"
)

func init() {
	registry.RegisterExtension(new(Factory), "consul")
}

type Factory struct {}

func (f *Factory) New(c *config.Config, uri *url.URL) registry.RegistryAdapter {
	config := consulapi.DefaultConfig()
	if uri.Host != "" {
		config.Address = uri.Host
	}

	if c.RegistryToken != "" {
		log.Printf("setting token to %s", c.RegistryToken)
		config.Token = c.RegistryToken
	}

	if c.RegistrySSL.Enabled {
		log.Printf("enabling SSL")
		config.Scheme = "https"
	}

	if !c.RegistrySSL.Verify {
		log.Printf("disabled SSL verification")
		config.HttpClient.Transport = &http.Transport {
			TLSClientConfig: &tls.Config {
				InsecureSkipVerify: true,
			},
		}
	}

	if c.RegistryAuth.Enabled {
		log.Printf("setting basic auth")
		config.HttpAuth = &consulapi.HttpBasicAuth{
			Username: c.RegistryAuth.Username,
			Password: c.RegistryAuth.Password,
		}
	}

	client, err := consulapi.NewClient(config)
	if err != nil {
		log.Fatal("consul: ", uri.Scheme)
	}
	return &ConsulAdapter{client: client}
}

type ConsulAdapter struct {
	client *consulapi.Client
}

func (r *ConsulAdapter) Register(service *registry.Service) error {
	registration := new(consulapi.AgentServiceRegistration)
	registration.ID		= service.ID
	registration.Name	= service.Name
	registration.Port	= service.Port
	registration.Tags	= service.Tags
	registration.Address	= service.IP
	return r.client.Agent().ServiceRegister(registration)
}

func (r *ConsulAdapter) Deregister(service *registry.Service) error {
	return r.client.Agent().ServiceDeregister(service.ID)
}
