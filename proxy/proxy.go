package proxy

import (
	"crypto/tls"
	"github.com/pawski/proxkeep/logger"
	"io/ioutil"
	"net/http"
	"net/url"
)

func Fetch(host, port, testURL string) (int, string, error) {

	logger.Get().Info("Proxy health check")
	transport := http.Transport{}
	transport.Proxy = http.ProxyURL(&url.URL{Scheme: "http", Host: host + ":" + port})
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //set ssl

	client := &http.Client{}
	client.Transport = &transport

	response, err := client.Get(testURL)

	if err != nil {
		logger.Get().Error(err)
		return 0, "", err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		logger.Get().Error(err)
		return 0, "", err
	}

	return response.StatusCode, string(body), nil
}

func DirectFetch(url string) (int, string, error) {
	client := &http.Client{}

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Get().Error(err)
		return 0, "", err
	}

	response, err := client.Do(request)

	if err != nil {
		logger.Get().Error(err)
		return 0, "", err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		logger.Get().Error(err)
		return 0, "", err
	}

	return response.StatusCode, string(body), nil
}
