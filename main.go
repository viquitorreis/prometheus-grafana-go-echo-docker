package main

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	customCounterVec *prometheus.CounterVec
	histogramVec     *prometheus.HistogramVec
)

func main() {
	e := echo.New()

	initPrometheusRequestCounter()
	initPrometheusHistogramVec()
	// adicionando o middleware do prometheus para monitorar as métricas por rota
	// vai monitorar a quantidade de requests por rota
	e.Use(prometheusRequestCounterMiddleware)
	e.Use(newEchoHistogramMiddleware)

	e.Use(middleware.BasicAuthWithConfig(middleware.BasicAuthConfig{
		Skipper: func(c echo.Context) bool {
			// para qualquer outra rota vai fazer skip da autenticação
			if c.Path() != "/metrics" {
				return true
			}

			return false
		},
		Validator: func(username, password string, c echo.Context) (bool, error) {
			// verifica se o usuário e senha são iguais a admin
			// caso seja, retorna true e vai para a rota /metrics
			// usar o subtle ao invés do operador == para evitar ataques de timing
			if subtle.ConstantTimeCompare([]byte(username), []byte("admin")) == 1 &&
				subtle.ConstantTimeCompare([]byte(password), []byte("admin")) == 1 {
				return true, nil
			}

			return false, nil
		},
	}))

	// adicionando o middleware do prometheus para monitorar as métricas por rota
	// e.Use(echoprometheus.NewMiddleware("tupa"))

	e.GET("/hello", func(c echo.Context) error {
		randValue := rand.Intn(101) + 100
		time.Sleep(time.Duration(randValue) * time.Millisecond)

		return c.String(http.StatusOK, fmt.Sprintf("Hello, World! generated in: %dms", randValue))
	})

	e.GET("/world", func(c echo.Context) error {
		randValue := rand.Intn(101) + 300
		time.Sleep(time.Duration(randValue) * time.Millisecond)

		return c.String(http.StatusOK, fmt.Sprintf("World! generated in: %dms", randValue))
	})

	e.GET("/metrics", echoprometheus.NewHandler())

	if err := e.Start(":8081"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Failed to start server: %v", err)
	}
}

var (
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_request_duration_seconds",
			Help: "Histogram of response latency (seconds) of requests grouped by endpoint.",
			// O DefBuckets automaticamente faz o split das durações em um bucket com presets padrão
			// como [0.005, 0.01, 0.025, 0.05, 0.1, ...] segundos. Pode ser customizado se quiser.
			Buckets: []float64{0.05, 0.1, 0.2, 0.5, 1, 2.5, 5, 10},
		},
		// label por path
		[]string{"path"},
	)
)

func init() {
	// registrando métrica customizada com o prometheus
	prometheus.MustRegister(requestDuration)
}

func newEchoHandlerWithHistogram(handler echo.HandlerFunc, histogram *prometheus.HistogramVec) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		path := c.Path()
		err := handler(c) // chamada para a rota - processa a request
		duration := time.Since(start).Seconds()
		respStatus := fmt.Sprintf("%d", c.Response().Status)

		// adiciona a duração da request ao histograma
		histogram.WithLabelValues(respStatus, path).Observe(duration)

		// OBS ------ ADICIONAR OS MÉTODOS HTTP TAMBÉM

		return err
	}
}

func newEchoHistogramMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		path := c.Path()
		err := next(c)
		duration := time.Since(start).Seconds()
		respStatus := fmt.Sprintf("%d", c.Response().Status)

		// adiciona a duração da request ao histograma
		histogramVec.WithLabelValues(respStatus, path).Observe(duration)

		return err
	}
}

func initPrometheusRequestCounter() {
	customCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_total_requests",
			Help: "Total number of HTTP requests per endpoint",
		},
		// endpoint de label
		[]string{"endpoint"},
	)

	if err := prometheus.Register(customCounterVec); err != nil {
		log.Fatal(fmt.Printf("Failed to register custom metrics: %v", err))
	}
}

func prometheusRequestCounterMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// incrementa o contador de requests
		customCounterVec.WithLabelValues(c.Path()).Inc()
		return next(c)
	}
}

func initPrometheusHistogramVec() {
	histogramVec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "prom_request_time",
		Help:    "Time it has taken to retrieve the metrics",
		Buckets: prometheus.DefBuckets,
	}, []string{"time", "endpoint"})

	// registrando histograma
	if err := prometheus.Register(histogramVec); err != nil {
		log.Fatal(fmt.Printf("Failed to register custom metrics: %v", err))
	}
}
