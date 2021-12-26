package health

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type HealthHandler struct{}

func New() *HealthHandler {
	return &HealthHandler{}
}

func GetName() string {
	return "health"
}

func (h *HealthHandler) GetHealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, "ok")
}
