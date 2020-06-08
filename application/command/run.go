package command

import (
	"context"
	"errors"
	"fmt"
	"github.com/pawski/proxkeep/application/stats"
	"github.com/pawski/proxkeep/domain/proxy"
	"net/http"
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
	statsServer      *http.Server
	totalToProcess   int64
	processedOk      *stats.Stats
	processedNok     *stats.Stats
}

func NewRunCommand(proxyTester *proxy.Tester, repository proxy.ReadWriteRepository, logger proxy.Logger) *RunCommand {
	return &RunCommand{proxyRepository: repository, logger: logger, proxyTester: proxyTester, stopFeedingQueue: false}
}

func (c *RunCommand) Execute(testURL string, maxConcurrentChecks uint) error {

	c.processedOk = &stats.Stats{}
	c.processedNok = &stats.Stats{}

	serveMux := http.NewServeMux()
	c.statsServer = &http.Server{Addr: ":8000", Handler: serveMux}

	serveMux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		// Math between total ok and nok might moth math under certain conditions (separate counters updated in different point in time), but that is ok at atm.
		okCnt := c.processedOk.Count()
		nokCnt := c.processedNok.Count()
		fmt.Fprintf(writer, "Processed servers: %v, OK: %v, NOK: %v. Remaining: %v", okCnt+nokCnt, okCnt, nokCnt, c.totalToProcess-(okCnt+nokCnt))
	})

	c.statsServer.RegisterOnShutdown(func() {
		c.logger.Info("Http stats server closed")
		c.wg.Done()
	})

	go func() {
		err := c.statsServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			c.logger.Fatal(err)
		}
	}()

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
	c.totalToProcess = int64(len(proxyList))
	c.logger.Infof("Proxies on list to check %v", c.totalToProcess)

	for _, v := range proxyList {
		if c.stopFeedingQueue {
			break
		}
		workloadQueue <- v
	}

	close(workloadQueue)

	c.wg.Wait()

	c.wg.Add(1)
	err = c.statsServer.Shutdown(context.Background())
	c.wg.Wait()

	if err != nil {
		return err
	}

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

	checkReport := c.proxyTester.Check(server.Ip, server.Port, test)

	if checkReport.ProxyOperational {
		c.logger.Infof("%v OK", server.Uid)
		c.logger.Infof("%v throughput %.3f KB/s", server.Uid, checkReport.ThroughputRate)
		c.processedOk.Add()
	} else {
		c.logger.Infof("%v NOK", server.Uid)
		c.logger.Debugf("%v failure reason %v", server.Uid, checkReport.FailureReason)
		c.processedNok.Add()
	}

	server.IsAvailable = checkReport.ProxyOperational
	server.FailureReason = checkReport.FailureReason
	server.ThroughputRate = float64(checkReport.ThroughputRate)

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
