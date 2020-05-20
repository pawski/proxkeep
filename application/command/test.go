package command

import (
	"errors"
	"github.com/pawski/proxkeep/domain/proxy"
)

type TestCommand struct {
	logger     proxy.Logger
	httpClient proxy.HttpClient
}

func NewTestCommand(httpClient proxy.HttpClient, logger proxy.Logger) *TestCommand {
	return &TestCommand{logger: logger, httpClient: httpClient}
}

func (c *TestCommand) Execute(testURL, ip, port string) error {
	if "" == port {
		port = "8080"
	}

	if "" == ip {
		c.logger.Errorf("IP Address cannot be empty")

		return nil
	}

	c.logger.Infof("Using %v:%v", ip, port)

	response, err := c.httpClient.DirectFetch(testURL)

	if err != nil {
		return err
	}

	if response.StatusCode != 200 {
		return errors.New("test URL returned non 200 response code")
	}

	c.logger.Info("Test data acquired")
	c.logger.Infof("Main connection Throughput %.3f KB/s", response.KiloBytesThroughputRate())

	test := proxy.NewResponseTest(testURL, response.StatusCode, response.Body)

	pResponse, _ := c.httpClient.Fetch(ip, port, test.GetTestURL())

	if test.Passed(pResponse.StatusCode, pResponse.Body) {
		c.logger.Info("Proxy - OK")
		c.logger.Infof("Proxy throughput %.3f KB/s", pResponse.KiloBytesThroughputRate())
	} else {
		c.logger.Info("Proxy - NOK")
	}

	return nil
}
