//go:generate go-extpoints . AdapterFactory
package registry

import (
	"net/url"
	"time"

	"github.com/CiscoCloud/mesos-consul/config"
)

type Config struct {
	MesosURI	string
	RegistryURI	string
	Refresh		time.Duration
}

type ServiceCheck struct {
	Script		string
	Interval	string
	Timeout		string
	TTL		string
	HTTP		string
}

type Service struct {
	ID	string
	Name	string
	Port	int
	IP	string
	Tags	[]string
	Check	*ServiceCheck
}

type AdapterFactory interface {
	New(c *config.Config, uri *url.URL) RegistryAdapter
}

type RegistryAdapter interface {
	Register(service *Service) error
	Deregister(server *Service) error
}
