package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func GetHelloPeople(c echo.Context) error {
	randValue := rand.Intn(101) + 100
	time.Sleep(time.Duration(randValue) * time.Millisecond)

	return c.String(http.StatusOK, fmt.Sprintf("[PEOPLE] Hello, World! generated in: %dms", randValue))
}

func GetWorldPeople(c echo.Context) error {
	randValue := rand.Intn(101) + 300
	time.Sleep(time.Duration(randValue) * time.Millisecond)

	return c.String(http.StatusOK, fmt.Sprintf("World! generated in: %dms", randValue))
}

func GetHelloAnimals(c echo.Context) error {
	randValue := rand.Intn(101) + 200
	time.Sleep(time.Duration(randValue) * time.Millisecond)

	return c.String(http.StatusOK, fmt.Sprintf("[ANIMALS] Hello, World! generated in: %dms", randValue))
}

func GetWorldAnimals(c echo.Context) error {
	randValue := rand.Intn(101) + 400
	time.Sleep(time.Duration(randValue) * time.Millisecond)

	return c.String(http.StatusOK, fmt.Sprintf("World! generated in: %dms", randValue))
}
