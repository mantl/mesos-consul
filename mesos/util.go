package mesos

import (
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"
)

func cleanName(name string) string {
	reg, err := regexp.Compile("[^\\w-.\\.]")
	if err != nil {
		log.Print(err)
		return name
	}

	s := reg.ReplaceAllString(name, "")

	return strings.ToLower(strings.Replace(s, "_", "", -1))
}

func leaderIP(leader string) string {
	host := strings.Split(leader, "@")[1]
	host = strings.Split(host, ":")[0]

	return host
}

func toIP(host string) string {
	ip, err := net.LookupIP(host)
	if err != nil {
		return host
	}

	return ip[0].String()
}

func toPort(p string) int {
	ps, err := strconv.Atoi(p)
	if err != nil {
		log.Printf("Invalid port number: %d", p)
	}

	return ps
}
