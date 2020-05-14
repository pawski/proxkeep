package command

import (
	"bytes"
	"errors"
	"github.com/pawski/proxkeep/infrastructure/repository"
	"github.com/pawski/proxkeep/logger"
	"github.com/pawski/proxkeep/proxy"
	"runtime"
	"sync"
)

type RunCommand struct {
	logger          logger.Logger
	proxyRepository *repository.ProxyServerRepository
}

func NewRunCommand(repository *repository.ProxyServerRepository, logger logger.Logger) *RunCommand {
	return &RunCommand{proxyRepository: repository, logger: logger}
}

func (c *RunCommand) Execute() error {
	var wg sync.WaitGroup

	c.logger.Infof("Max parallel checks: %v", runtime.GOMAXPROCS(-1))

	testUrl := "https://letsencrypt.org/documents/LE-SA-v1.2-November-15-2017.pdf"

	response, err := proxy.DirectFetch(testUrl)

	if response.StatusCode != 200 {
		return errors.New("test URL returned non 200 response code")
	}

	if err != nil {
		return err
	}

	c.logger.Info("Test data acquired")
	c.logger.Infof("Main connection Throughput %.4f KB/s", float64(len(response.Body)/1024)/response.TransferTime)

	for _, v := range c.proxyRepository.FindAll() {

		wg.Add(1)
		go func(v repository.ProxyServer) {
			c.logger.Infof("Using %v:%v", v.Ip, v.Port)
			pResponse, _ := proxy.Fetch(v.Ip, string(v.Port), testUrl)

			if response.StatusCode == pResponse.StatusCode && 0 == bytes.Compare(response.Body, pResponse.Body) {
				c.logger.Info("Proxy - OK")
				c.logger.Infof("Proxy throughput %.4f KB/s", float64(len(pResponse.Body)/1024)/pResponse.TransferTime)
			} else {
				c.logger.Info("Proxy - NOK")
			}

			wg.Done()
		}(v)
	}

	wg.Wait()

	return nil
}
