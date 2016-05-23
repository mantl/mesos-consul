package mesos

import (
	"regexp"
	"strings"

	"github.com/CiscoCloud/mesos-consul/registry"
	"github.com/CiscoCloud/mesos-consul/state"
)

type CheckVar struct {
	Host string
	Port string
}

var globalCV *CheckVar

const defaultInterval = "2s"
const defaultTimeout = "60s"

// Task Methods

// GetCheck()
//   Build a Check structure from the Task labels
//
func GetCheck(t *state.Task, cv *CheckVar) *registry.Check {
	matched := false
	c := registry.DefaultCheck()
	for _, l := range t.Labels {
		k := strings.ToLower(l.Key)

		switch k {
		case "check_http":
			c.HTTP = interpolate(cv, l.Value)
			matched = true
		case "check_script":
			c.Script = interpolate(cv, l.Value)
			matched = true
		case "check_ttl":
			c.TTL = interpolate(cv, l.Value)
			matched = true
		case "check_interval":
			c.Interval = l.Value
			matched = true
		}
	}

	if !matched {
		c.TCP = interpolate(cv, hostPortToString(cv))
		c.Interval = defaultInterval
		c.Timeout = defaultTimeout

	}

	return c
}

// Returns a CheckVar struct in "IP:PORT" format
//
func hostPortToString(cv *CheckVar) string {
	strResult := ""
	if cv.Host != "" && cv.Port != "" {
		strResult = strings.Join([]string{cv.Host, ":", cv.Port}, "")
	}

	return strResult
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
}
