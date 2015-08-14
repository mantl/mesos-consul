package mesos

import (
	"fmt"
	"strings"

	"github.com/CiscoCloud/mesos-consul/registry"
)

func (sj *StateJSON) GetFollowerById(id string) (string, error) {
	for _, f := range sj.Followers {

		if f.Id == id {
			return f.Hostname, nil
		}
	}

	return "", fmt.Errorf("Follower not found: %s", id)
}

// Task Methods

// GetCheck()
//   Build a Check structure from the Task labels
//
func (t *Task) GetCheck() *registry.Check {
	c := registry.DefaultCheck()

	for _, l := range t.Labels {
		k := strings.ToLower(l.Key)

		switch k {
		case "consul_http_check":
			c.HTTP = l.Value
		case "consul_script_check":
			c.Script = l.Value
		case "consul_ttl_check":
			c.TTL = l.Value
		case "consul_check_interval":
			c.Interval = l.Value
		}
	}

	return c
}
