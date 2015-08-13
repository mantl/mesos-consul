package consul

import (
	"log"
	"strings"

	"github.com/CiscoCloud/mesos-consul/registry"

	consulapi	"github.com/hashicorp/consul/api"
)

type cacheEntry struct {
	service		*consulapi.AgentServiceRegistration
	isRegistered	bool
}

// Service cache
var serviceCache map[string]*cacheEntry

// CacheCreate()
//
func (c *Consul) CacheCreate() {
	if serviceCache == nil {
		serviceCache = make(map[string]*cacheEntry)
	}
}

// Initialize the service cache
//
func (c *Consul) CacheLoad(host string) error {
	client := c.client(host).Catalog()

	serviceList, _, err := client.Services(nil)
	if err != nil {
		return err
	}

	for service, _ := range serviceList {
		catalogServices, _, err := client.Service(service, "", nil)
		if err != nil {
			return err
		}

		for _, s := range catalogServices {
			if strings.HasPrefix(s.ServiceID, "mesos-consul:")  {
				log.Printf("[DEBUG] Found '%s' with ID '%s'", s.ServiceName, s.ServiceID)
				serviceCache[s.ServiceID] = &cacheEntry{
					service:	&consulapi.AgentServiceRegistration{
							ID:		s.ServiceID,
							Name:		s.ServiceName,
							Port:		s.ServicePort,
							Address:	s.ServiceAddress,
							Tags:		s.ServiceTags,
							},
					isRegistered:   false,
				}
			}
		}
	}

	return nil
}

// CacheLookup()
//
func (c *Consul) CacheLookup(id string) *registry.Service {
	if _, ok := serviceCache[id]; ok {
		s := serviceCache[id].service

		return &registry.Service{
			ID:		s.ID,
			Name:		s.Name,
			Port:		s.Port,
			Address:	s.Address,
			Tags:		s.Tags,
		}
	}

	return nil
}

// CacheDelete()
//
func (c *Consul) CacheDelete(id string) {
	if _, ok := serviceCache[id]; ok {
		delete(serviceCache, id)
	}
}

// CacheMark()
//   Mark the service ID as registered
//
func (c *Consul) CacheMark(id string) {
	if _, ok := serviceCache[id]; ok {
		serviceCache[id].isRegistered = true
	}
}
