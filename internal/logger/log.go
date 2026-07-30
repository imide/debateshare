package logger

import (
	"os"
	"strings"
	"time"

	"github.com/imide/debateshare/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Init(cfg *config.Config, silent bool) {
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
		NoColor:    false,
		FormatLevel: func(i any) string {
			if ll, ok := i.(string); ok {
				switch ll {
				case "trace":
					return "\033[37m" + strings.ToUpper(ll) + "\033[0m" // White/Gray
				case "debug":
					return "\033[36m" + strings.ToUpper(ll) + "\033[0m" // Cyan
				case "info":
					return "\033[32m" + strings.ToUpper(ll) + "\033[0m" // Green
				case "warn":
					return "\033[33m" + strings.ToUpper(ll) + "\033[0m" // Yellow
				case "error":
					return "\033[31m" + strings.ToUpper(ll) + "\033[0m" // Red
				case "fatal":
					return "\033[35m" + strings.ToUpper(ll) + "\033[0m" // Magenta
				case "panic":
					return "\033[41m\033[37m" + strings.ToUpper(ll) + "\033[0m" // White on Red background
				default:
					return strings.ToUpper(ll)
				}
			}
			return "???"
		},
	}

	log.Logger = zerolog.New(output).
		With().
		Timestamp().
		Caller().
		Logger()

	// get logger level from config, default to "info"
	logLevel := strings.ToLower(cfg.Log.Level)

	// set log level based on config
	switch logLevel {
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "warn", "warning":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	default:
		// Use GinMode from server config instead of logging config
		if cfg.App.Debug || os.Getenv("APP_ENV") == "dev" {
			zerolog.SetGlobalLevel(zerolog.TraceLevel)
			if !silent {
				log.Debug().Msg("Logger initialized in debug mode")
			}
		} else {
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		}
	}
}

func Get() zerolog.Logger {
	return log.Logger
}
