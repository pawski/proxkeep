package command

import (
	"errors"
	"fmt"
	"github.com/pawski/proxkeep/application/service"
	"github.com/pawski/proxkeep/application/stats"
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
	measurement      *service.Measurement
}

var bus *stats.EventBus
var testService *service.ProxyTester

func NewRunCommand(proxyTester *proxy.Tester, repository proxy.ReadWriteRepository, logger proxy.Logger, m *service.Measurement) *RunCommand {
	return &RunCommand{proxyRepository: repository, logger: logger, proxyTester: proxyTester, stopFeedingQueue: false, measurement: m}
}

func (c *RunCommand) Execute(testURL string, maxConcurrentChecks uint) error {
	bus = stats.NewEventBus()
	c.measurement.StartHTTP()

	pOkSubscriber := make(stats.Subscriber)
	pNokSubscriber := make(stats.Subscriber)
	pTotalSubscriber := make(stats.Subscriber)

	go func(subscriber stats.Subscriber, measurement *service.Measurement) {
		for _ = range subscriber {
			measurement.AddOk()
		}
	}(pOkSubscriber, c.measurement)

	go func(subscriber stats.Subscriber, measurement *service.Measurement) {
		for _ = range subscriber {
			measurement.AddNok()
		}
	}(pNokSubscriber, c.measurement)

	go func(subscriber stats.Subscriber, measurement *service.Measurement) {
		for event := range subscriber {
			measurement.SetTotal(event.Data.(int64))
		}
	}(pTotalSubscriber, c.measurement)

	bus.Subscribe(stats.ProcessedOk, pOkSubscriber)
	bus.Subscribe(stats.ProcessedNok, pNokSubscriber)
	bus.Subscribe(stats.TotalToProcess, pTotalSubscriber)

	if "" == testURL {
		return errors.New("testURL cannot be empty string")
	}

	if 0 == maxConcurrentChecks {
		return errors.New("macConcurrentChecks can't be zero")
	}
	c.logger.Infof("Max concurrent checks %v", maxConcurrentChecks)

	testService = service.NewProxyTester(c.proxyTester, c.proxyRepository, c.logger, bus)
	c.stopFeedingOnSigTerm()

	err := testService.Run(testURL, maxConcurrentChecks)

	if err != nil {
		return err
	}

	return nil
}

func (c *RunCommand) stopFeedingOnSigTerm() {
	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt)
	signal.Notify(s, syscall.SIGTERM)

	go func(testService *service.ProxyTester) {
		<-s
		fmt.Println("Feeding queue stopped, waiting for remaining jobs...")
		testService.GracefulShutdown()
	}(testService)
}
