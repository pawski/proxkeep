package service

import (
	"github.com/influxdata/influxdb/client/v2"
	"github.com/pawski/proxkeep/application/configuration"
	"github.com/pawski/proxkeep/application/stats"
	"github.com/pawski/proxkeep/domain/proxy"
	"sync"
	"time"
)

type MetricsCollector struct {
	eventBus                  *stats.EventBus
	processed                 *stats.Stats
	logger                    proxy.Logger
	influx                    client.Client
	bp                        client.BatchPoints
	metricsWriteInterval      time.Duration
	metricsCollectionInterval time.Duration
}

func NewMetricsCollector(bus *stats.EventBus, l proxy.Logger) *MetricsCollector {
	return &MetricsCollector{
		eventBus:                  bus,
		logger:                    l,
		processed:                 &stats.Stats{},
		metricsWriteInterval:      time.Second * 10, // make configurable as app.yml
		metricsCollectionInterval: time.Second,      // make configurable as app.yml
	}
}

func (m *MetricsCollector) SubscribeOk() {

	pOkSubscriber := make(stats.Subscriber)

	go func(subscriber stats.Subscriber, measurement *MetricsCollector) {
		for _ = range subscriber {
			measurement.addOk()
		}
	}(pOkSubscriber, m)

	m.eventBus.Subscribe(stats.ProcessedOk, pOkSubscriber)
}

func (m *MetricsCollector) addOk() {
	m.processed.Add()
}

func (m *MetricsCollector) addNok() {
	m.processed.Add()
}

func (m *MetricsCollector) SubscribeNok() {
	pNokSubscriber := make(stats.Subscriber)

	go func(subscriber stats.Subscriber, measurement *MetricsCollector) {
		for _ = range subscriber {
			measurement.addNok()
		}
	}(pNokSubscriber, m)

	m.eventBus.Subscribe(stats.ProcessedNok, pNokSubscriber)
}

func (m *MetricsCollector) StartMonitor() {
	env, err := configuration.GetEnv()

	if err != nil {
		m.logger.Fatal(err)
	}

	influx, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     env.InfluxDbHost,
		Username: env.InfluxDbUser,
		Password: env.InfluxDbPassword,
	})

	if err != nil {
		m.logger.Fatal(err)
	}

	flushTicker := time.NewTicker(m.metricsWriteInterval)

	m.bp, err = client.NewBatchPoints(client.BatchPointsConfig{
		Database:  env.InfluxDbDatabase,
		Precision: "s",
	})

	if err != nil {
		m.logger.Fatal(err)
	}

	mux := sync.Mutex{}

	go func(mux *sync.Mutex) {
		for _ = range flushTicker.C {
			if len(m.bp.Points()) > 0 {
				mux.Lock()
				err = influx.Write(m.bp)

				if err != nil {
					m.logger.Fatal(err)
				} else {
					m.bp, _ = client.NewBatchPoints(client.BatchPointsConfig{
						Database:  env.InfluxDbDatabase,
						Precision: "s",
					})
				}
				mux.Unlock()
			}
		}
	}(&mux)

	buffTicker := time.NewTicker(m.metricsCollectionInterval)

	go func(mux *sync.Mutex) {
		for _ = range buffTicker.C {
			pt, err := client.NewPoint("runtime", map[string]string{}, map[string]interface{}{"count": m.processed.Count()}, time.Now())

			if err != nil {
				m.logger.Error(err)
			}

			mux.Lock()
			m.bp.AddPoint(pt)
			mux.Unlock()
		}
	}(&mux)
}
