package mesos

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/CiscoCloud/mesos-consul/registry"
)

type CheckVar struct {
	Host string
	Port string
}
var globalCV *CheckVar

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
func (t *Task) GetCheck(cv *CheckVar) *registry.Check {
	c := registry.DefaultCheck()

	for _, l := range t.Labels {
		k := strings.ToLower(l.Key)

		switch k {
		case "check_http":
			c.HTTP = interpolate(cv, l.Value)
		case "check_script":
			c.Script = interpolate(cv, l.Value)
		case "check_ttl":
			c.TTL = interpolate(cv, l.Value)
		case "check_interval":
			c.Interval = l.Value
		}
	}

	return c
}

// Replace {variables} with values
//
func interpolate(cv *CheckVar, s string) string {
	r := regexp.MustCompile("{[^}]*}")

	globalCV = cv
	rval := r.ReplaceAllStringFunc(s, varReplace)

	return string(rval)
}

// Replacement function
//
func varReplace(s string) string {
	switch s {
	case "{port}":
		return globalCV.Port
	case "{host}":
		return globalCV.Host
	default:
		return s
	}

	return s
}
