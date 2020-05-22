package http

import (
	"crypto/tls"
	"github.com/pawski/proxkeep/domain/proxy"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type httpClient struct {
	httpConnectionTimeout uint
	logger                proxy.Logger
}

func NewHttpClient(connectionTimeout uint, logger proxy.Logger) *httpClient {
	return &httpClient{httpConnectionTimeout: connectionTimeout, logger: logger}
}

func (h *httpClient) Fetch(host, port, testURL string) (*proxy.HttpResponse, error) {

	transport := http.Transport{}
	transport.Proxy = http.ProxyURL(&url.URL{Scheme: "http", Host: host + ":" + port})
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	client := http.Client{
		Timeout: time.Second * time.Duration(h.httpConnectionTimeout),
	}
	client.Transport = &transport

	start := time.Now()
	response, err := client.Get(testURL)
	duration := time.Since(start).Seconds()

	if err != nil {
		h.logger.Debug(err)
		return &proxy.HttpResponse{StatusCode: 0, Body: []byte{}, TransferTime: duration}, err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		h.logger.Debug(err)
		return &proxy.HttpResponse{StatusCode: 0, Body: []byte{}, TransferTime: duration}, err
	}

	return &proxy.HttpResponse{StatusCode: response.StatusCode, Body: body, TransferTime: duration}, nil
}

func (h *httpClient) DirectFetch(url string) (proxy.HttpResponse, error) {
	client := http.Client{}

	request, err := http.NewRequest("GET", url, nil)

	if err != nil {
		h.logger.Debug(err)
		return proxy.HttpResponse{StatusCode: 0, Body: []byte{}, TransferTime: 0}, err
	}

	start := time.Now()
	response, err := client.Do(request)
	duration := time.Since(start).Seconds()

	if err != nil {
		h.logger.Debug(err)
		return proxy.HttpResponse{StatusCode: 0, Body: []byte{}, TransferTime: duration}, err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		h.logger.Debug(err)
		return proxy.HttpResponse{StatusCode: 0, Body: []byte{}, TransferTime: duration}, err
	}

	return proxy.HttpResponse{StatusCode: response.StatusCode, Body: body, TransferTime: duration}, nil
}
