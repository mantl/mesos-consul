package mesos

import (
	"log"
	"regexp"
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
