package service

import (
	"github.com/pawski/proxkeep/application/stats"
	"github.com/pawski/proxkeep/domain/proxy"
	"sync"
)

type ProxyTester struct {
	proxyTester      *proxy.Tester
	proxyRepository  proxy.ReadWriteRepository
	logger           proxy.Logger
	wg               sync.WaitGroup
	stopFeedingQueue bool
	events           *stats.EventBus
}

func NewProxyTester(proxyTester *proxy.Tester, repository proxy.ReadWriteRepository, logger proxy.Logger, bus *stats.EventBus) *ProxyTester {
	return &ProxyTester{proxyRepository: repository, logger: logger, proxyTester: proxyTester, stopFeedingQueue: false, events: bus}
}

func (c *ProxyTester) Run(testURL string, maxConcurrentChecks uint) error {

	c.logger.Infof("Testing using %v", testURL)
	proxyTest, err := c.proxyTester.PrepareTest(testURL)

	if err != nil {
		return err
	}

	c.logger.Info("Test data acquired")

	workloadQueue := make(chan proxy.Server)
	semaphore := make(chan struct{}, maxConcurrentChecks)

	c.wg.Add(1)
	go c.dispatchWorkload(workloadQueue, semaphore, proxyTest)

	proxyList := c.proxyRepository.FindAll()
	totalToProcess := int64(len(proxyList))

	c.events.Publish(stats.EventData{Topic: stats.TotalToProcess, Data: totalToProcess})

	c.logger.Infof("Proxies on list to check %v", totalToProcess)

	for _, v := range proxyList {
		if c.stopFeedingQueue {
			break
		}
		workloadQueue <- v
	}

	close(workloadQueue)

	c.wg.Wait()

	return nil
}

func (c *ProxyTester) dispatchWorkload(workQueue <-chan proxy.Server, semaphore chan struct{}, test *proxy.ResponseTest) {
	defer c.wg.Done()
	for v := range workQueue {
		semaphore <- struct{}{}
		c.wg.Add(1)
		go c.work(v, semaphore, test)
	}
}

func (c *ProxyTester) work(server proxy.Server, sem <-chan struct{}, test *proxy.ResponseTest) {
	defer func() { <-sem }()
	defer c.wg.Done()

	checkReport := c.proxyTester.Check(server.Ip, server.Port, test)

	if checkReport.ProxyOperational {
		c.logger.Debugf("%v OK", server.Uid)
		c.logger.Infof("%v Throughput %.3f KB/s", server.Uid, checkReport.ThroughputRate)
		c.events.Publish(stats.EventData{Topic: stats.ProcessedOk, Data: struct{}{}})
	} else {
		c.logger.Debugf("%v NOK", server.Uid)
		c.logger.Debugf("%v Failure reason %v", server.Uid, checkReport.FailureReason)
		c.events.Publish(stats.EventData{Topic: stats.ProcessedNok, Data: struct{}{}})
	}

	server.IsAvailable = checkReport.ProxyOperational
	server.FailureReason = checkReport.FailureReason
	server.ThroughputRate = float64(checkReport.ThroughputRate)

	err := c.proxyRepository.Persist(server)
	if err != nil {
		c.logger.Errorf("%v %v", server.Uid, err)
	}
}

func (c *ProxyTester) GracefulShutdown() {
	c.stopFeedingQueue = true
}
