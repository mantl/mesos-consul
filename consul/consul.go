package consul

import (
	"log"
	"net/url"

	"github.com/CiscoCloud/mesos-consul/registry"
	consulapi "github.com/hashicorp/consul/api"
)

func init() {
	registry.RegisterExtension(new(Factory), "consul")
}

type Factory struct {}

func (f *Factory) New(uri *url.URL) registry.RegistryAdapter {
	config := consulapi.DefaultConfig()
	if uri.Host != "" {
		config.Address = uri.Host
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
