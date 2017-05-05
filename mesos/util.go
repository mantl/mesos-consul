package mesos

import (
	"net"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

func cleanName(name string, separator string, groupPrefixes []string) string {
	_name := name
	for _, p := range(groupPrefixes) {
		reg, err := regexp.Compile(p)
		if err != nil {
			log.Warn(err)
			continue
		}

		if groups := reg.FindStringSubmatch(_name); len(groups) > 1 {
			_name = groups[1]
			break
		}
	}
	reg, err := regexp.Compile("[^\\w-]")
	if err != nil {
		log.Warn(err)
		return _name
	}

	s := reg.ReplaceAllString(_name, "-")

	return strings.ToLower(strings.Replace(s, "_", separator, -1))
}

// helper function to compare service tag slices
//
func sliceEq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func sliceContainsString(s []string, b string) bool {
	for _, a := range s {
		if a == b {
			return true
		}
	}
	return false
}

func leaderIP(leader string) string {
	host := strings.Split(leader, "@")[1]
	host = strings.Split(host, ":")[0]

	return toIP(host)
}

func toIP(host string) string {
	// Check if host string is already an IP address
	ip := net.ParseIP(host)
	if ip != nil {
		return host
	}

	// Try to resolve host
	i, err := net.ResolveIPAddr("ip4", host)
	if err != nil {
		// Return the hostname if unable to resolve
		return host
	}

	return i.String()
}

func toPort(p string) int {
	ps, err := strconv.Atoi(p)
	if err != nil {
		log.Warnf("Invalid port number: %s", p)
	}

	return ps
}
