package mesos

import (
	"testing"
)

func TestLeaderIP(t *testing.T) {
	l := "master@124.123.123.121:5050"

	ip := leaderIP(l)

	t.Log("ip: ", ip)
}

func TestParsePID(t *testing.T) {
	l := "slave(1)@127.0.0.1:5051"

	host, port := parsePid(l)

	t.Log("host: ", host)
}
