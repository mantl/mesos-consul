package mesos

import (
	"testing"
)

func TestLeaderIP(t *testing.T) {
	l := "master@124.123.123.121:5050"

	ip := leaderIP(l)

	t.Log("ip: ", ip)
}
