package command

import (
	"bytes"
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

func (c *TestCommand) Execute(ip, port string) error {
	if "" == port {
		port = "8080"
	}

	if "" == ip {
		c.logger.Errorf("IP Address cannot be empty")

		return nil
	}

	c.logger.Infof("Using %v:%v", ip, port)

	testUrl := "https://letsencrypt.org/documents/LE-SA-v1.2-November-15-2017.pdf"
	response, err := http.DirectFetch(testUrl)

	if response.StatusCode != 200 {
		return errors.New("test URL returned non 200 response code")
	}

	if err != nil {
		return err
	}

	c.logger.Info("Test data acquired")
	c.logger.Infof("Main connection Throughput %.4f KB/s", float64(len(response.Body)/1024)/response.TransferTime)

	pResponse, _ := http.Fetch(ip, port, testUrl)

	if response.StatusCode == pResponse.StatusCode && 0 == bytes.Compare(response.Body, pResponse.Body) {
		c.logger.Info("Proxy - OK")
		c.logger.Infof("Proxy throughput %.4f KB/s", float64(len(pResponse.Body)/1024)/pResponse.TransferTime)
	} else {
		c.logger.Info("Proxy - NOK")
	}

	return nil
}
