package mesos

type Follower struct {
	Id       string `json:"id"`
	Hostname string `json:"hostname"`
	Pid      string `json:"pid"`
}

type Resources struct {
	Ports string `json:"ports"`
}

type Label struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Task struct {
	FrameworkId string  `json:"framework_id"`
	Id          string  `json:"id"`
	Name        string  `json:"name"`
	FollowerId  string  `json:"slave_id"`
	State       string  `json:"state"`
	Labels      []Label `json:"labels"`
	Resources   `json:"resources"`
}

type Framework struct {
	Tasks []Task `json:"tasks"`
	Name  string `json:"name"`
}

type StateJSON struct {
	Frameworks []Framework `json:"frameworks"`
	Followers  []Follower  `json:"slaves"`
	Leader     string      `json:"leader"`
}

type MesosHost struct {
	host         string
	port         string
	isLeader     bool
	isRegistered bool
}
