package logger

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		if path == "/health" {
			return
		}

		var event *zerolog.Event
		switch {
		case c.Writer.Status() >= 500:
			event = log.Error()
		case c.Writer.Status() >= 400:
			event = log.Warn()
		default:
			event = log.Info()
		}

		event.
			Str("method", c.Request.Method).
			Str("path", path).
			Int("status", c.Writer.Status()).
			Dur("latency", time.Since(start))

		if query != "" {
			event.Str("query", query)
		}

		if len(c.Errors) > 0 {
			event.Str("error", c.Errors.String())
		}

		if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
			event.Str("request_id", requestID)
		}
	}
}
