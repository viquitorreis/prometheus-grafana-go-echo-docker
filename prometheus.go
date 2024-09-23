package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	reqCounterVecs   = make(map[string]*prometheus.CounterVec)
	reqCounterVecsMU sync.RWMutex
)

var (
	reqDurationHistogramVecs   = make(map[string]*prometheus.HistogramVec)
	reqDurationHistogramVecsMU sync.RWMutex
)

type CustomContextKey string

const (
	SchemaKey CustomContextKey = "customPrometheusContext"
)

func prometheusContextMiddleware(schema string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(string(SchemaKey), schema)

			initPrometheusRequestCounter(schema)
			initPromReqDurationHistogramVec(schema)

			return next(c)
		}
	}
}

func initPrometheusRequestCounter(schema string) {
	name := fmt.Sprintf("http_request_total_%s", schema)

	requestCounterVec := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: name,
			Help: "Total number of HTTP requests per endpoint",
		},
		// endpoint de label
		[]string{"endpoint"},
	)

	if err := prometheus.Register(requestCounterVec); err != nil {
		log.Fatal(fmt.Printf("Failed to register custom metrics: %v", err))
	}

	// adiciona o contador ao map e schema específico enquanto está em mutex
	reqCounterVecsMU.Lock()
	defer reqCounterVecsMU.Unlock()
	reqCounterVecs[schema] = requestCounterVec
}

func prometheusRequestCounterMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// se a rota for /metrics, não incrementa o contador
		if c.Path() == "/metrics" {
			return next(c)
		}

		schema := c.Get(string(SchemaKey)).(string)
		println(fmt.Sprintf("prometheusRequestCounterMiddleware schema: %s", schema))

		// precisamos de um mutex para garantir que o map vai ser lido de forma segura
		reqCounterVecsMU.RLock()
		counterVec, ok := reqCounterVecs[schema]
		reqCounterVecsMU.RUnlock()

		if ok {
			counterVec.WithLabelValues(c.Path()).Inc()
		} else {
			log.Printf("Failed to increment request counter for schema %s", schema)
		}

		return next(c)
	}
}

func initPromReqDurationHistogramVec(schema string) {
	name := fmt.Sprintf("http_request_duration_seconds_%s", schema)

	reqDurationHistogramVec := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    name,
		Help:    "Time it has taken to retrieve the metrics",
		Buckets: prometheus.DefBuckets,
	}, []string{"schema", "endpoint", "status"})

	// registrando histograma
	if err := prometheus.Register(reqDurationHistogramVec); err != nil {
		log.Fatal(fmt.Printf("Failed to register custom metrics: %v", err))
	}

	// adiciona o contador ao map e schema específico enquanto está em mutex
	reqDurationHistogramVecsMU.Lock()
	defer reqDurationHistogramVecsMU.Unlock()
	reqDurationHistogramVecs[schema] = reqDurationHistogramVec
}

func prometheusHistogramReqDurationMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		path := c.Path()
		err := next(c)
		duration := time.Since(start).Seconds()
		respStatus := fmt.Sprintf("%d", c.Response().Status)

		schema := c.Get(string(SchemaKey))
		println(fmt.Sprintf("prometheusHistogramReqDurationMiddleware schema: %s", schema))

		// precisamos de um mutex para garantir que o map vai ser lido de forma segura
		reqCounterVecsMU.RLock()
		histogramVec, ok := reqDurationHistogramVecs[schema.(string)]
		reqCounterVecsMU.RUnlock()

		if ok {
			// adiciona a duração da request ao histograma
			histogramVec.WithLabelValues(schema.(string), path, respStatus).Observe(duration)
		} else {
			log.Printf("Failed to update request duration histogram for schema %s", path)
		}

		return err
	}
}
