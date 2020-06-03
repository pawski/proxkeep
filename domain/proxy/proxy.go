package proxy

type Uid string

type CheckReport struct {
	ProxyIdentifier  string
	ProxyOperational bool
	ThroughputRate   ThroughputRate
	FailureReason    string
}

type ThroughputRate float64

func (r *HttpResponse) KiloBytesThroughputRate() float64 {
	return float64(len(r.Body)/1000) / r.TransferTime
}
