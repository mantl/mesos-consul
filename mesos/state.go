package mesos

import (
	"fmt"

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
	return registry.DefaultCheck()
}
