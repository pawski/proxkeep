package proxy

import (
	"errors"
	"math"
)

type Tester struct {
	httpClient HttpClient
}

func NewTester(httpClient HttpClient) *Tester {
	return &Tester{httpClient: httpClient}
}

func (t *Tester) PrepareTest(testURL string) (*ResponseTest, error) {

	response, err := t.httpClient.DirectFetch(testURL)

	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, errors.New("test URL returned non 200 expectedResponse code")
	}

	return NewResponseTest(testURL, response.StatusCode, response.Body), nil
}

func (t *Tester) Check(host string, port string, test *ResponseTest) CheckReport {
	var proxyReport = CheckReport{ProxyIdentifier: host + ":" + port}

	pResponse, pError := t.httpClient.Fetch(host, port, test.GetTestURL())

	if pError == nil && test.Passed(pResponse.StatusCode, pResponse.Body) {
		proxyReport.ProxyOperational = true
		proxyReport.ThroughputRate = ThroughputRate(math.Round(pResponse.KiloBytesThroughputRate()*1000) / 1000)
	} else {
		proxyReport.ProxyOperational = false
		proxyReport.ThroughputRate = ThroughputRate(0)
		if pError != nil {
			proxyReport.FailureReason = pError.Error()
		}
	}

	return proxyReport
}
