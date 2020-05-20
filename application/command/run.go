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

var stopFeedingQueue = false

type RunCommand struct {
	logger          proxy.Logger
	proxyRepository *repository.ProxyServerRepository
}

func NewRunCommand(repository *repository.ProxyServerRepository, logger proxy.Logger) *RunCommand {
	return &RunCommand{proxyRepository: repository, logger: logger}
}

func (c *RunCommand) Execute(testURL string, maxConcurrentChecks uint) error {
	var wg sync.WaitGroup

	if "" == testURL {
		return errors.New("testURL cannot be empty string")
	}

	if 0 == maxConcurrentChecks {
		return errors.New("macConcurrentChecks can't be zero")
	}

	setupTerminationOnSigTerm()

	c.logger.Infof("Max concurrent checks %v", maxConcurrentChecks)

	expectedResponse, err := http.DirectFetch(testURL)

	if expectedResponse.StatusCode != 200 {
		return errors.New("test URL returned non 200 expectedResponse code")
	}

	if err != nil {
		return err
	}

	c.logger.Infof("Testing using %v", testURL)
	c.logger.Info("Test data acquired")
	c.logger.Infof("Main connection Throughput %.3f KB/s", math.Round((expectedResponse.BytesThroughputRate()/1024)*1000)/1000)

	proxyTest := proxy.Prepare(testURL, expectedResponse.StatusCode, expectedResponse.Body)
	workloadQueue := make(chan repository.ServerEntity)
	semaphore := make(chan struct{}, maxConcurrentChecks)

	wg.Add(1)
	go func(workQueue <-chan repository.ServerEntity, wg *sync.WaitGroup, semaphore chan struct{}) {
		defer (*wg).Done()
		for v := range workQueue {
			semaphore <- struct{}{}
			(*wg).Add(1)
			go c.work(v, semaphore, wg, proxyTest)
		}
	}(workloadQueue, &wg, semaphore)

	proxyList := c.proxyRepository.FindAll()
	c.logger.Infof("Proxies on list to check %v", len(proxyList))

	for _, v := range proxyList {
		if stopFeedingQueue {
			break
		}
		workloadQueue <- v
	}

	close(workloadQueue)

	wg.Wait()

	return nil
}

func (c *RunCommand) work(server repository.ServerEntity, sem <-chan struct{}, wg *sync.WaitGroup, test *proxy.ResponseTest) {
	defer func() { <-sem }()
	defer wg.Done()

	c.logger.Debugf("%v test", server.Uid)

	checkReport := check(server.Ip, server.Port, test)

	if checkReport.ProxyOperational {
		c.logger.Infof("%v OK", server.Uid)
		c.logger.Infof("%v throughput %.3f KB/s", server.Uid, checkReport.ThroughputRate)
	} else {
		c.logger.Infof("%v NOK", server.Uid)
		c.logger.Debugf("%v failure reason %v", server.Uid, checkReport.FailureReason)
	}

	server.IsAvailable = checkReport.ProxyOperational
	server.FailureReason = checkReport.FailureReason
	server.ThroughputRate = checkReport.ThroughputRate.AsBytes()

	err := c.proxyRepository.Persist(&server)
	if err != nil {
		c.logger.Errorf("%v %v", server.Uid, err)
	}
}

func check(host string, port string, proxyTest *proxy.ResponseTest) *proxy.CheckReport {

	var proxyReport = proxy.CheckReport{ProxyIdentifier: host + ":" + port}

	pResponse, pError := http.Fetch(host, port, proxyTest.GetTestURL())

	if pError == nil && proxyTest.Passed(pResponse.StatusCode, pResponse.Body) {
		proxyReport.ProxyOperational = true
		proxyReport.ThroughputRate = proxy.ThroughputRate(math.Round((pResponse.BytesThroughputRate()/1024)*1000) / 1000)
	} else {
		proxyReport.ProxyOperational = false
		proxyReport.ThroughputRate = proxy.ThroughputRate(0)
		if pError != nil {
			proxyReport.FailureReason = pError.Error()
		}
	}

	return &proxyReport
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
