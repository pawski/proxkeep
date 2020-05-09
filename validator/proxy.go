package validator

import (
	"crypto/tls"
	"github.com/pawski/go-xchange/logger"
	"io/ioutil"
	"net/http"
	"net/url"
)

func HealthCheck(host, port, expectedResponse string) {

	logger.Get().Info("Proxy health check")
	transport := http.Transport{}
	transport.Proxy = http.ProxyURL(&url.URL{Scheme: "http", Host: host + ":" + port})
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //set ssl

	client := &http.Client{}
	client.Transport = &transport

	response, err := client.Get("https://letsencrypt.org/documents/LE-SA-v1.2-November-15-2017.pdf")

	if err != nil {
		logger.Get().Error(err)
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		logger.Get().Error(err)
	}

	logger.Get().Info(response.StatusCode)
	logger.Get().Info(len(body))
	logger.Get().Info(len(expectedResponse))
}

// ifconfig.io/all.json
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
