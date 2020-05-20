package command

import (
	"errors"
	"fmt"
	"github.com/pawski/proxkeep/domain/proxy"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type RunCommand struct {
	proxyTester      *proxy.Tester
	proxyRepository  proxy.ReadWriteRepository
	logger           proxy.Logger
	wg               sync.WaitGroup
	stopFeedingQueue bool
}

func NewRunCommand(proxyTester *proxy.Tester, repository proxy.ReadWriteRepository, logger proxy.Logger) *RunCommand {
	return &RunCommand{proxyRepository: repository, logger: logger, proxyTester: proxyTester, stopFeedingQueue: false}
}

func (c *RunCommand) Execute(testURL string, maxConcurrentChecks uint) error {
	if "" == testURL {
		return errors.New("testURL cannot be empty string")
	}

	if 0 == maxConcurrentChecks {
		return errors.New("macConcurrentChecks can't be zero")
	}
	c.logger.Infof("Max concurrent checks %v", maxConcurrentChecks)

	c.stopFeedingOnSigTerm()

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
	c.logger.Infof("Proxies on list to check %v", len(proxyList))

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

func (c *RunCommand) dispatchWorkload(workQueue <-chan proxy.Server, semaphore chan struct{}, test *proxy.ResponseTest) {
	defer c.wg.Done()
	for v := range workQueue {
		semaphore <- struct{}{}
		c.wg.Add(1)
		go c.work(v, semaphore, test)
	}
}

func (c *RunCommand) work(server proxy.Server, sem <-chan struct{}, test *proxy.ResponseTest) {
	defer func() { <-sem }()
	defer c.wg.Done()

	c.logger.Debugf("%v test", server.Uid)

	checkReport := c.proxyTester.Check(server.Ip, server.Port, test)

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

func (c *RunCommand) stopFeedingOnSigTerm() {
	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt)
	signal.Notify(s, syscall.SIGTERM)
	go func(stopFeeding *bool) {
		<-s
		fmt.Println("Feeding queue stopped, waiting for remaining jobs...")
		*stopFeeding = true
	}(&c.stopFeedingQueue)
}
