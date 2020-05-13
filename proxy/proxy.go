package proxy

import (
	"crypto/tls"
	"github.com/pawski/proxkeep/logger"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type Response struct {
	Body         []byte
	StatusCode   int
	TransferTime float64
}

func Fetch(host, port, testURL string) (Response, error) {

	logger.Get().Info("Proxy health check")
	transport := http.Transport{}
	transport.Proxy = http.ProxyURL(&url.URL{Scheme: "http", Host: host + ":" + port})
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	client := &http.Client{}
	client.Transport = &transport

	start := time.Now()
	response, err := client.Get(testURL)
	duration := time.Since(start).Seconds()

	if err != nil {
		logger.Get().Error(err)
		return Response{StatusCode: 0, Body: []byte{}, TransferTime: duration}, err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		logger.Get().Error(err)
		return Response{StatusCode: response.StatusCode, Body: []byte{}, TransferTime: duration}, err
	}

	return Response{StatusCode: response.StatusCode, Body: body, TransferTime: duration}, nil
}

func DirectFetch(url string) (Response, error) {
	client := &http.Client{}

	request, err := http.NewRequest("GET", url, nil)

	if err != nil {
		logger.Get().Error(err)
		return Response{StatusCode: 0, Body: []byte{}, TransferTime: 0}, err
	}

	start := time.Now()
	response, err := client.Do(request)
	duration := time.Since(start).Seconds()

	if err != nil {
		logger.Get().Error(err)
		return Response{StatusCode: 0, Body: []byte{}, TransferTime: duration}, err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		logger.Get().Error(err)
		return Response{StatusCode: response.StatusCode, Body: []byte{}, TransferTime: duration}, err
	}

	return Response{StatusCode: response.StatusCode, Body: body, TransferTime: duration}, nil
}
