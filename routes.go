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
)

func initECHO() {
	e := echo.New()

	e.Use(prometheusRequestCounterMiddleware)
	e.Use(prometheusHistogramReqDurationMiddleware)

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

	e.GET("/metrics", echoprometheus.NewHandler())

	initAllRouteManagers(e)

	if err := e.Start(":8081"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func initAllRouteManagers(e *echo.Echo) {
	peopleRouteManager(e)
	animalsRouteManager(e)
}

func peopleRouteManager(e *echo.Echo) {
	schema := "people"

	g := e.Group(fmt.Sprintf("/%s", schema))
	// g.Use(prometheusRequestCounterMiddleware(schema))
	// g.Use(prometheusHistogramReqDurationMiddleware(schema))
	g.Use(prometheusContextMiddleware(schema))

	g.GET("/hello", GetHelloPeople)
	g.GET("/world", GetWorldPeople)
}

func animalsRouteManager(e *echo.Echo) {
	schema := "animals"

	initPrometheusRequestCounter(schema)
	initPromReqDurationHistogramVec(schema)

	g := e.Group(fmt.Sprintf("/%s", schema))
	// g.Use(prometheusRequestCounterMiddleware(schema))
	// g.Use(prometheusHistogramReqDurationMiddleware(schema))
	g.Use(prometheusContextMiddleware(schema))

	g.GET("/hello", GetHelloAnimals)
	g.GET("/world", GetWorldAnimals)
}
