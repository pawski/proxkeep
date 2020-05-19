package http

import (
	"crypto/tls"
	"github.com/pawski/proxkeep/infrastructure/logger/logrus"
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

func (r Response) BytesThroughputRate() float64 {
	return float64(len(r.Body)) / r.TransferTime
}

func Fetch(host, port, testURL string) (Response, error) {

	transport := http.Transport{}
	transport.Proxy = http.ProxyURL(&url.URL{Scheme: "http", Host: host + ":" + port})
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	client := &http.Client{
		Timeout: time.Second * 10,
	}
	client.Transport = &transport

	start := time.Now()
	response, err := client.Get(testURL)
	duration := time.Since(start).Seconds()

	if err != nil {
		logrus.Get().Debug(err)
		return Response{StatusCode: 0, Body: []byte{}, TransferTime: duration}, err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		logrus.Get().Debug(err)
		return Response{StatusCode: 0, Body: []byte{}, TransferTime: duration}, err
	}

	return Response{StatusCode: response.StatusCode, Body: body, TransferTime: duration}, nil
}

func DirectFetch(url string) (Response, error) {
	client := &http.Client{}

	request, err := http.NewRequest("GET", url, nil)

	if err != nil {
		logrus.Get().Debug(err)
		return Response{StatusCode: 0, Body: []byte{}, TransferTime: 0}, err
	}

	start := time.Now()
	response, err := client.Do(request)
	duration := time.Since(start).Seconds()

	if err != nil {
		logrus.Get().Debug(err)
		return Response{StatusCode: 0, Body: []byte{}, TransferTime: duration}, err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		logrus.Get().Debug(err)
		return Response{StatusCode: 0, Body: []byte{}, TransferTime: duration}, err
	}

	return Response{StatusCode: response.StatusCode, Body: body, TransferTime: duration}, nil
}
