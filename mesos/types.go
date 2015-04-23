package mesos

type follower struct {
	Id		string	`json:"id"`
	Hostname	string	`json:"hostname"`
	Pid		string	`json:"pid"`
}

type Followers []follower

type Resources struct {
	Ports		string	`json:"ports"`
}

type Tasks []struct {
	FrameworkId	string	`json:"framework_id"`
	Id		string	`json:"id"`
	Name		string	`json:"name"`
	FollowerId	string	`json:"slave_id"`
	State		string	`json:"state"`
	Resources		`json:"resources"`
}

type Frameworks []struct {
	Tasks			`json:"tasks"`
	Name		string	`json:"name"`
}

type StateJSON struct {
	Frameworks		`json:"frameworks"`
	Followers		`json:"slaves"`
	Leader		string	`json:"leader"`
}

type MesosHost struct {
	host		string
	port		string
	isLeader	bool
	isRegistered	bool
}
