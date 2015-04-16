package registry

import (
	"log"
	"net/url"
)

func GetRegistry(RegistryURI string) RegistryAdapter {
	uri, err := url.Parse(RegistryURI)
	if err != nil {
		log.Fatal("Bad registry URI: ", RegistryURI)
	}

	factory := AdapterFactories.Lookup(uri.Scheme)
	if factory == nil {
		log.Fatal("Unrecognized registry: ", RegistryURI)
	}

	registry := factory.New(uri)

	return registry
}

