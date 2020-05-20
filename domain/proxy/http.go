package proxy

type HttpResponse struct {
	Body         []byte
	StatusCode   int
	TransferTime float64
}

type HttpClient interface {
	DirectFetch(url string) (HttpResponse, error)
	Fetch(host, port, testURL string) (*HttpResponse, error)
}
