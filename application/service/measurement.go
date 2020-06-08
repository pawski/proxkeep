package service

import (
	"context"
	"fmt"
	"github.com/pawski/proxkeep/application/stats"
	"github.com/pawski/proxkeep/domain/proxy"
	"net/http"
	"sync"
)

type Measurement struct {
	wg          sync.WaitGroup
	statsServer *http.Server
	logger      proxy.Logger
	values      measurements
}

type measurements struct {
	processedOk    *stats.Stats
	processedNok   *stats.Stats
	totalToProcess int64
}

func NewMeasurement(l proxy.Logger) *Measurement {
	return &Measurement{logger: l, values: measurements{
		processedOk:    &stats.Stats{},
		processedNok:   &stats.Stats{},
		totalToProcess: 0,
	}}
}

func (m *Measurement) AddOk() {
	m.values.processedOk.Add()
}

func (m *Measurement) AddNok() {
	m.values.processedNok.Add()
}

func (m *Measurement) SetTotal(total int64) {
	fmt.Print("EHLO!")
	m.values.totalToProcess = total
}

func (m *Measurement) StartHTTP() {
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

func (m *Measurement) StopHTTP() error {
	m.wg.Add(1)
	err := m.statsServer.Shutdown(context.Background())
	m.wg.Wait()

	return err
}
