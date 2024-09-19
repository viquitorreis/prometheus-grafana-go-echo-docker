package main

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	e := echo.New()
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

	customCounter := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "custom_requests_total",
			Help: "Number of HTTP requests processed by Tupã",
		},
	)
	if err := prometheus.Register(customCounter); err != nil {
		log.Fatal(fmt.Printf("Failed to register custom metrics: %v", err))
	}
	// rota para o prometheus, aqui colocamos o nome da aplicação
	e.Use(echoprometheus.NewMiddleware("tupa"))

	e.GET("/hello", func(c echo.Context) error {
		customCounter.Inc()
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.GET("/metrics", echoprometheus.NewHandler())

	if err := e.Start(":8081"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Failed to start server: %v", err)
	}
}
