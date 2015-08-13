package registry

type Check struct {
	Script		string
	TTL		string
	HTTP		string
	Interval	string
}

type Service struct {
	ID		string
	Name		string
	Port		int
	Address		string
	Tags		[]string
	Check		*Check
}

type Registry interface {
	CacheCreate()
	CacheDelete(string)
	CacheLoad(string) error
	CacheLookup(string) *Service
	CacheMark(string)

	Register(*Service) error
	Deregister() error
}
