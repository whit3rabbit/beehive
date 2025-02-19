package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/whit3rabbit/beehive/manager/internal/logger"
	"go.uber.org/zap"
)

// CustomHTTPErrorHandler handles errors in a consistent way across the application
func CustomHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	message := "Internal Server Error"

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		message = he.Message.(string)
	}

	logger.Error("HTTP error",
		zap.Int("status_code", code),
		zap.String("error", message),
		zap.String("path", c.Path()))

	if !c.Response().Committed {
		if c.Request().Method == http.MethodHead {
			err = c.NoContent(code)
		} else {
			err = c.JSON(code, map[string]string{"error": message})
		}
		if err != nil {
			logger.Error("Failed to send error response", zap.Error(err))
		}
	}
}
