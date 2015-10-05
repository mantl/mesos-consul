package mesos

import (
	"regexp"
	"strings"

	"github.com/CiscoCloud/mesos-consul/registry"

	"github.com/mesosphere/mesos-dns/records/state"
)

type CheckVar struct {
	Host string
	Port string
}
var globalCV *CheckVar

// Task Methods

// GetCheck()
//   Build a Check structure from the Task labels
//
func GetCheck(t *state.Task, cv *CheckVar) *registry.Check {
	c := registry.DefaultCheck()

	for _, l := range t.DiscoveryInfo.Labels.Labels {
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
