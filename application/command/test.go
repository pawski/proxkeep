package command

import (
	"errors"
	"github.com/pawski/proxkeep/domain/proxy"
	"github.com/pawski/proxkeep/infrastructure/network/http"
)

type TestCommand struct {
	logger proxy.Logger
}

func NewTestCommand(logger proxy.Logger) *TestCommand {
	return &TestCommand{logger: logger}
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

	response, err := http.DirectFetch(testURL)

	if response.StatusCode != 200 {
		return errors.New("test URL returned non 200 response code")
	}

	if err != nil {
		return err
	}

	c.logger.Info("Test data acquired")
	c.logger.Infof("Main connection Throughput %.3f KB/s", response.KiloBytesThroughputRate())

	test := proxy.Prepare(testURL, response.StatusCode, response.Body)

	pResponse, _ := http.Fetch(ip, port, test.GetTestURL())

	if test.Passed(pResponse.StatusCode, pResponse.Body) {
		c.logger.Info("Proxy - OK")
		c.logger.Infof("Proxy throughput %.3f KB/s", pResponse.KiloBytesThroughputRate())
	} else {
		c.logger.Info("Proxy - NOK")
	}

	return nil
}
