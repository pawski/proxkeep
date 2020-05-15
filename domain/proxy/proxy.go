package proxy

type Uid string

type ServerEntity struct {
	Id             uint
	Uid            string
	Ip             string
	Port           string
	IsAvailable    bool
	ThroughputRate float64
	FailureReason  string
	CreatedAt      string
	UpdatedAt      string
}
