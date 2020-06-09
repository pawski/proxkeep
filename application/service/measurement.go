package service

import (
	"context"
	"fmt"
	"github.com/pawski/proxkeep/application/stats"
	"github.com/pawski/proxkeep/domain/proxy"
	"net/http"
	"sync"
)

type MeasurementService struct {
	wg          sync.WaitGroup
	statsServer *http.Server
	logger      proxy.Logger
	eventBus    *stats.EventBus
	values      measurements
}

type measurements struct {
	processedOk    *stats.Stats
	processedNok   *stats.Stats
	totalToProcess int64
}

func NewMeasurement(bus *stats.EventBus, l proxy.Logger) *MeasurementService {
	return &MeasurementService{eventBus: bus, logger: l, values: measurements{
		processedOk:    &stats.Stats{},
		processedNok:   &stats.Stats{},
		totalToProcess: 0,
	}}
}

func (m *MeasurementService) SubscribeOk() {

	pOkSubscriber := make(stats.Subscriber)

	go func(subscriber stats.Subscriber, measurement *MeasurementService) {
		for _ = range subscriber {
			measurement.addOk()
		}
	}(pOkSubscriber, m)

	m.eventBus.Subscribe(stats.ProcessedOk, pOkSubscriber)
}

func (m *MeasurementService) addOk() {
	m.values.processedOk.Add()
}

func (m *MeasurementService) addNok() {
	m.values.processedNok.Add()
}

func (m *MeasurementService) SubscribeNok() {
	pNokSubscriber := make(stats.Subscriber)

	go func(subscriber stats.Subscriber, measurement *MeasurementService) {
		for _ = range subscriber {
			measurement.addNok()
		}
	}(pNokSubscriber, m)

	m.eventBus.Subscribe(stats.ProcessedNok, pNokSubscriber)
}

func (m *MeasurementService) setTotal(total int64) {
	m.values.totalToProcess = total
}

func (m *MeasurementService) SubscribeTotal() {
	pTotalSubscriber := make(stats.Subscriber)

	go func(subscriber stats.Subscriber, measurement *MeasurementService) {
		for event := range subscriber {
			measurement.setTotal(event.Data.(int64))
		}
	}(pTotalSubscriber, m)

	m.eventBus.Subscribe(stats.TotalToProcess, pTotalSubscriber)
}

func (m *MeasurementService) StartHTTP() {
	serveMux := http.NewServeMux()
	m.statsServer = &http.Server{Addr: ":8000", Handler: serveMux}

	serveMux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		okCnt := m.values.processedOk.Count()
		nokCnt := m.values.processedNok.Count()
		fmt.Fprintf(writer, "Processed servers: %v, OK: %v, NOK: %v. Remaining: %v", okCnt+nokCnt, okCnt, nokCnt, m.values.totalToProcess-(okCnt+nokCnt))
	})

	m.statsServer.RegisterOnShutdown(func() {
		m.logger.Info("Http stats server closed")
		m.wg.Done()
	})

	go func() {
		err := m.statsServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			m.logger.Fatal(err)
		}
	}()
}

func (m *MeasurementService) GracefulShutdownHTTP() error {
	m.wg.Add(1)
	err := m.statsServer.Shutdown(context.Background())
	m.wg.Wait()

	return err
}
