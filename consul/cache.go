package consul

import (
	"strings"

	"github.com/CiscoCloud/mesos-consul/registry"

	consulapi "github.com/hashicorp/consul/api"
	log "github.com/sirupsen/logrus"
)

type cacheEntry struct {
	service         *consulapi.AgentServiceRegistration
	agent           string
	validityCounter int
}

func newCacheEntry(service *consulapi.AgentServiceRegistration, agent string) *cacheEntry {
	return &cacheEntry{
		agent:           service.Address,
		service:         service,
		validityCounter: 0,
	}
}

// Service cache
var serviceCache map[string]*cacheEntry
var cacheEntryValidityThreshold int = 1

// CacheCreate()
//
func (c *Consul) CacheCreate() bool {
	if serviceCache == nil {
		serviceCache = make(map[string]*cacheEntry)
		return true
	}

	return false
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
			if strings.HasPrefix(s.ServiceID, "mesos-consul:") {
				log.Debugf("Found '%s' with ID '%s'", s.ServiceName, s.ServiceID)
				serviceCache[s.ServiceID] = newCacheEntry(&consulapi.AgentServiceRegistration{
					ID:      s.ServiceID,
					Name:    s.ServiceName,
					Port:    s.ServicePort,
					Address: s.ServiceAddress,
					Tags:    s.ServiceTags,
				}, s.Address)
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
			ID:      s.ID,
			Name:    s.Name,
			Port:    s.Port,
			Address: s.Address,
			Tags:    s.Tags,
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
//   Mark the service ID as valid
//
func (c *Consul) CacheMark(id string) {
	if _, ok := serviceCache[id]; ok {
		serviceCache[id].validityCounter = 0
	}
}

// CacheProcessDeregister()
//   Calculate the validity of the entry
//
func (c *Consul) CacheProcessDeregister(id string) {
	if _, ok := serviceCache[id]; ok {
		serviceCache[id].validityCounter++
	}
}

func (c *Consul) CacheIsValid(id string) bool {
	if _, ok := serviceCache[id]; ok {
		return serviceCache[id].validityCounter < cacheEntryValidityThreshold
	}
	return false
}
