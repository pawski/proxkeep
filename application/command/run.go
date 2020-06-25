package command

import (
	"errors"
	"github.com/pawski/proxkeep/application/service"
	"github.com/pawski/proxkeep/domain/proxy"
	"os"
	"os/signal"
	"syscall"
)

type RunCommand struct {
	testService             *service.ProxyTester
	measurementService      *service.MeasurementService
	metricsCollectorService *service.MetricsCollector
	logger                  proxy.Logger
}

func NewRunCommand(testService *service.ProxyTester, logger proxy.Logger, m *service.MeasurementService, mCollector *service.MetricsCollector) *RunCommand {
	return &RunCommand{logger: logger, measurementService: m, testService: testService, metricsCollectorService: mCollector}
}

func (c *RunCommand) Execute(testURL string, maxConcurrentChecks uint, metricsCollectorEnabled bool, httpStatsEnabled bool) error {

	if "" == testURL {
		return errors.New("testURL cannot be empty string")
	}

	if 0 == maxConcurrentChecks {
		return errors.New("maxConcurrentChecks can't be zero")
	}
	c.logger.Infof("Max concurrent checks %v", maxConcurrentChecks)

	if metricsCollectorEnabled {
		c.metricsCollectorService.SubscribeNok()
		c.metricsCollectorService.SubscribeOk()
		c.metricsCollectorService.StartMonitor()
	}

	if httpStatsEnabled {
		c.measurementService.SubscribeOk()
		c.measurementService.SubscribeNok()
		c.measurementService.SubscribeTotal()

		c.logger.Info()
		c.measurementService.StartHTTP()
	}

	c.stopFeedingOnSigTerm()
	err := c.testService.Run(testURL, maxConcurrentChecks)

	if err != nil {
		return err
	}

	if httpStatsEnabled {
		err = c.measurementService.GracefulShutdownHTTP()

		if err != nil {
			return err
		}
	}

	return nil
}

func (c *RunCommand) stopFeedingOnSigTerm() {
	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt)
	signal.Notify(s, syscall.SIGTERM)

	go func(testService *service.ProxyTester) {
		<-s
		c.logger.Infof("Feeding queue stopped, waiting for remaining jobs...")
		testService.GracefulShutdown()
	}(c.testService)
}
