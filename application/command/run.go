package command

import (
	"errors"
	"fmt"
	"github.com/pawski/proxkeep/domain/proxy"
	"github.com/pawski/proxkeep/infrastructure/network/http"
	"github.com/pawski/proxkeep/infrastructure/repository"
	"math"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var maxConcurrentChecks = 20
var testUrl = "https://letsencrypt.org/documents/LE-SA-v1.2-November-15-2017.pdf"
var stopFeedingQueue = false

type RunCommand struct {
	logger          proxy.Logger
	proxyRepository *repository.ProxyServerRepository
}

func NewRunCommand(repository *repository.ProxyServerRepository, logger proxy.Logger) *RunCommand {
	return &RunCommand{proxyRepository: repository, logger: logger}
}

func (c *RunCommand) Execute() error {
	setupTerminationOnSigTerm()

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
	c.logger.Infof("Max concurrent checks %v", maxConcurrentChecks)

	proxyTest := proxy.Prepare(expectedResponse.StatusCode, expectedResponse.Body)
	workloadQueue := make(chan repository.ServerEntity)
	semaphore := make(chan struct{}, maxConcurrentChecks)

	wg.Add(1)
	go func(workQueue <-chan repository.ServerEntity, wg *sync.WaitGroup) {
		defer wg.Done()

		for v := range workQueue {
			wg.Add(1)
			semaphore <- struct{}{}
			go func(proxyServer repository.ServerEntity, semaphore <-chan struct{}, wg *sync.WaitGroup) {
				defer wg.Done()

				c.logger.Debugf("%v test", proxyServer.Uid)

				report := check(v.Ip, v.Port, testUrl, proxyTest)

				if report.ProxyOperational {
					c.logger.Infof("%v OK", proxyServer.Uid)
					c.logger.Infof("%v throughput %.3f KB/s", proxyServer.Uid, report.ThroughputRate)
				} else {
					c.logger.Infof("%v NOK", proxyServer.Uid)
					c.logger.Debugf("%v failure reason %v", proxyServer.Uid, report.FailureReason)
				}

				proxyServer.IsAvailable = report.ProxyOperational
				proxyServer.FailureReason = report.FailureReason
				proxyServer.ThroughputRate = report.ThroughputRate.AsBytes()

				c.proxyRepository.Persist(proxyServer)
				<-semaphore
			}(v, semaphore, wg)
		}
	}(workloadQueue, &wg)

	proxyList := c.proxyRepository.FindAll()
	c.logger.Infof("Proxies on list to check %v", len(proxyList))

	for _, v := range proxyList {
		workloadQueue <- v
		if stopFeedingQueue {
			break
		}
	}

	close(workloadQueue)

	wg.Wait()

	return nil
}

func check(host string, port string, testURL string, proxyTest *proxy.ResponseTest) *proxy.CheckReport {

	var proxyReport = &proxy.CheckReport{ProxyIdentifier: host + ":" + port}

	pResponse, pError := http.Fetch(host, port, testURL)

	if pError != nil && proxyTest.Passed(pResponse.StatusCode, pResponse.Body) {
		proxyReport.ProxyOperational = true
		proxyReport.ThroughputRate = proxy.ThroughputRate(math.Round((pResponse.BytesThroughputRate()/1024)*1000) / 1000)
	} else {
		proxyReport.ProxyOperational = false
		proxyReport.ThroughputRate = proxy.ThroughputRate(0)
		if pError != nil {
			proxyReport.FailureReason = pError.Error()
		}
	}

	return proxyReport
}

func setupTerminationOnSigTerm() {
	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt)
	signal.Notify(s, syscall.SIGTERM)
	go func() {
		<-s
		fmt.Println("Feeding queue stopped, waiting for remaining jobs...")
		stopFeedingQueue = true
	}()
}
