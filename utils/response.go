package utils

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func JSONError(context echo.Context, err error) error {
	// Log all errors
	context.Logger().Error(err)

	newErr, ok := err.(*Error)
	if !ok {
		context.Logger().Panic("error not wrapped: ", err)
		return context.JSON(http.StatusNotImplemented, err)
	}

	return context.JSON(newErr.Code, newErr)
}
