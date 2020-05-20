package proxy

type Uid string

type CheckReport struct {
	ProxyIdentifier  string
	ProxyOperational bool
	ThroughputRate   ThroughputRate
	FailureReason    string
}

type ThroughputRate float64

func (t ThroughputRate) AsKiloBytes() float64 {
	return t.AsBytes() / 1024
}

func (t ThroughputRate) AsBytes() float64 {
	return float64(t)
}

type HttpResponse struct {
	Body         []byte
	StatusCode   int
	TransferTime float64
}

func (r HttpResponse) KiloBytesThroughputRate() float64 {
	return float64(len(r.Body)/1000) / r.TransferTime
}
