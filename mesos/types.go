package mesos

import (
	"sync"
)

type MesosLeader struct {
	host		string
	port		string
	leaderLock	sync.Mutex
}

type slave struct {
	Id		string	`json:"id"`
	Hostname	string	`json:"hostname"`
}

type Slaves []slave

type Resources struct {
	Ports		string	`json:"ports"`
}

type Tasks []struct {
	FrameworkId	string	`json:"framework_id"`
	Id		string	`json:"id"`
	Name		string	`json:"name"`
	SlaveId		string	`json:"slave_id"`
	State		string	`json:"state"`
	Resources		`json:"resources"`
}

type Frameworks []struct {
	Tasks			`json:"tasks"`
	Name		string	`json:"name"`
}

type StateJSON struct {
	Frameworks		`json:"frameworks"`
	Slaves			`json:"slaves"`
	Leader		string	`json:"leader"`
}
