package command

import (
	"bytes"
	"errors"
	"github.com/pawski/proxkeep/domain/proxy"
	"github.com/pawski/proxkeep/infrastructure/network/http"
	"github.com/pawski/proxkeep/infrastructure/repository"
	"math"
	"sync"
)

var maxConcurrentChecks = 20
var testUrl = "https://letsencrypt.org/documents/LE-SA-v1.2-November-15-2017.pdf"

type RunCommand struct {
	logger          proxy.Logger
	proxyRepository *repository.ProxyServerRepository
}

func NewRunCommand(repository *repository.ProxyServerRepository, logger proxy.Logger) *RunCommand {
	return &RunCommand{proxyRepository: repository, logger: logger}
}

func (c *RunCommand) Execute() error {
	var wg sync.WaitGroup

	expectedResponse, err := http.DirectFetch(testUrl)

	if expectedResponse.StatusCode != 200 {
		return errors.New("test URL returned non 200 expectedResponse code")
	}

	if err != nil {
		return err
	}

	c.logger.Info("Test data acquired")
	c.logger.Infof("Main connection Throughput %.3f KB/s", math.Round((expectedResponse.BytesThroughputRate()/1024)*1000)/1000)

	workQueue := make(chan proxy.ServerEntity)
	semaphore := make(chan struct{}, maxConcurrentChecks)

	wg.Add(1)
	go func(workQueue <-chan proxy.ServerEntity, wg *sync.WaitGroup) {
		defer wg.Done()

		for v := range workQueue {
			wg.Add(1)
			semaphore <- struct{}{}
			go func(proxyServer proxy.ServerEntity, semaphore <-chan struct{}, wg *sync.WaitGroup) {
				defer wg.Done()

				c.logger.Infof("%v test started", proxyServer.Uid)

				pResponse, pError := http.Fetch(v.Ip, v.Port, testUrl)

				if expectedResponse.StatusCode == pResponse.StatusCode && 0 == bytes.Compare(expectedResponse.Body, pResponse.Body) {
					c.logger.Infof("%v OK", proxyServer.Uid)

					proxyServer.IsAvailable = true
					proxyServer.ThroughputRate = math.Round((pResponse.BytesThroughputRate()/1024)*1000) / 1000
					proxyServer.FailureReason = ""

					c.logger.Infof("%v throughput %.3f KB/s", proxyServer.Uid, proxyServer.ThroughputRate)
				} else {
					c.logger.Infof("%v NOK", proxyServer.Uid)

					proxyServer.IsAvailable = false
					proxyServer.ThroughputRate = 0
					proxyServer.FailureReason = ""
					if pError != nil {
						proxyServer.FailureReason = pError.Error()
					}
				}

				c.proxyRepository.Persist(proxyServer)
				<-semaphore
			}(v, semaphore, wg)
		}
	}(workQueue, &wg)

	for _, v := range c.proxyRepository.FindAll() {
		c.logger.Info("Add one")
		workQueue <- v
	}

	close(workQueue)

	wg.Wait()

	return nil
}
