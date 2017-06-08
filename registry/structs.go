package registry

type Check struct {
	Script   string
	TTL      string
	HTTP     string
	Interval string
	TCP      string
	Timeout  string
}

type Service struct {
	ID      string
	Name    string
	Port    int
	Address string
	Tags    []string
	Check   *Check
	Agent   string
}

type Registry interface {
	CacheCreate() bool
	CacheDelete(string)
	CacheLoad(string, string) error
	CacheLookup(string) *Service
	CacheMark(string)

	Register(*Service)
	Deregister()
}

func DefaultCheck() *Check {
	return &Check{
		TTL:      "",
		Script:   "",
		HTTP:     "",
		Interval: "",
		TCP:      "",
		Timeout:  "",
	}
}
