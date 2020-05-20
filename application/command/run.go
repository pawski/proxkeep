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

var maxConcurrentChecks = 100
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
	var wg sync.WaitGroup

	setupTerminationOnSigTerm()

	c.logger.Infof("Max concurrent checks %v", maxConcurrentChecks)

	expectedResponse, err := http.DirectFetch(testUrl)

	if expectedResponse.StatusCode != 200 {
		return errors.New("test URL returned non 200 expectedResponse code")
	}

	if err != nil {
		return err
	}

	c.logger.Info("Test data acquired")
	c.logger.Infof("Main connection Throughput %.3f KB/s", math.Round((expectedResponse.BytesThroughputRate()/1024)*1000)/1000)

	proxyTest := proxy.Prepare(expectedResponse.StatusCode, expectedResponse.Body)
	workloadQueue := make(chan repository.ServerEntity)
	semaphore := make(chan struct{}, maxConcurrentChecks)

	wg.Add(1)
	go func(workQueue <-chan repository.ServerEntity, wg *sync.WaitGroup) {
		defer (*wg).Done()
		for v := range workQueue {
			(*wg).Add(1)
			semaphore <- struct{}{}
			go c.work(v, semaphore, wg, proxyTest)
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

func (c *RunCommand) work(server repository.ServerEntity, sem <-chan struct{}, wg *sync.WaitGroup, test *proxy.ResponseTest) {
	defer func() { <-sem }()
	defer wg.Done()

	c.logger.Debugf("%v test", server.Uid)

	checkReport := check(server.Ip, server.Port, testUrl, test)

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

	err := c.proxyRepository.Persist(server)
	if err != nil {
		c.logger.Errorf("%v %v", server.Uid, err)
	}
}

func check(host string, port string, testURL string, proxyTest *proxy.ResponseTest) *proxy.CheckReport {

	var proxyReport = proxy.CheckReport{ProxyIdentifier: host + ":" + port}

	pResponse, pError := http.Fetch(host, port, testURL)

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
