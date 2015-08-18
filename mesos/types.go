package mesos

type follower struct {
	Id       string `json:"id"`
	Hostname string `json:"hostname"`
	Pid      string `json:"pid"`
}

type Followers []follower

type Label struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Resources struct {
	Ports string `json:"ports"`
}

type Status struct {
	Timestamp float64 `json:"timestamp"`
	State     string  `json:"state"`
	Labels    []Label `json:"labels,omitempty"`
}

type Task struct {
	FrameworkId string   `json:"framework_id"`
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	FollowerId  string   `json:"slave_id"`
	State       string   `json:"state"`
	Statuses    []Status `json:"statuses"`
	Resources   `json:"resources"`
}

type Frameworks []struct {
	Tasks []Task `json:"tasks"`
	Name  string `json:"name"`
}

type StateJSON struct {
	Frameworks `json:"frameworks"`
	Followers  `json:"slaves"`
	Leader     string `json:"leader"`
}

type MesosHost struct {
	host         string
	port         string
	isLeader     bool
	isRegistered bool
}
