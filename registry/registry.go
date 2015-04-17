package registry

import (
	"net/url"
	"log"

	"github.com/CiscoCloud/mesos-consul/config"
)

func GetRegistry(c *config.Config) RegistryAdapter {
	uri, err := url.Parse(c.Registry)
	if err != nil {
		log.Fatal("Bad registry URI: ", c.Registry)
	}

	factory := AdapterFactories.Lookup(uri.Scheme)
	if factory == nil {
		log.Fatal("Unrecognized registry: ", c.Registry)
	}

	registry := factory.New(c, uri)

	return registry
}

