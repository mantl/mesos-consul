package mesos

import (
	"fmt"
)

// Look up a follwer host name by the follower ID
func (fs *Followers) hostById(id string) (string, error) {
	for _, f := range *fs {
		if f.Id == id {
			return f.Hostname, nil
		}
	}

	return "", fmt.Errorf("Follower not found: %s", id)
}
