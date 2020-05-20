package proxy

type Finder interface {
	FindAll() []Server
	FindByUid(uid Uid) Server
}

type Persister interface {
	Persist(entity Server)
}

type Server struct {
	Uid            string
	Ip             string
	Port           string
	IsAvailable    bool
	ThroughputRate float64
	FailureReason  string
}
