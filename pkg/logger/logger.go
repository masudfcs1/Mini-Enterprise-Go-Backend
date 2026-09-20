package logger

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Log is the global zerolog instance.
var Log zerolog.Logger

// ANSI color escape codes
const (
	ColorReset   = "\033[0m"
	ColorBold    = "\033[1m"
	ColorRed     = "\033[31m"
	ColorGreen   = "\033[32m"
	ColorYellow  = "\033[33m"
	ColorBlue    = "\033[34m"
	ColorMagenta = "\033[35m"
	ColorCyan    = "\033[36m"
	ColorGray    = "\033[90m"
)

// InitLogger initializes the global logger based on environment.
func InitLogger(env string) {
	zerolog.TimeFieldFormat = time.RFC3339

	if env == "production" {
		Log = zerolog.New(os.Stdout).With().Timestamp().Logger()
	} else {
		// Pino-pretty style console writer with Windows colorable support
		output := zerolog.ConsoleWriter{
			Out:        colorable.NewColorableStdout(),
			TimeFormat: "15:04:05.000",
			NoColor:    false,
		}
		Log = zerolog.New(output).With().Timestamp().Logger()
	}

	log.Logger = Log
}

// HTTPLogger provides a Pino-style colorful HTTP request logger for Chi.
func HTTPLogger() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()

			defer func() {
				duration := time.Since(start)
				status := ww.Status()
				if status == 0 {
					status = http.StatusOK
				}

				statusColor := ColorGreen
				switch {
				case status >= 500:
					statusColor = ColorRed
				case status >= 400:
					statusColor = ColorYellow
				case status >= 300:
					statusColor = ColorCyan
				}

				methodColor := ColorCyan
				switch r.Method {
				case http.MethodGet:
					methodColor = ColorBlue
				case http.MethodPost:
					methodColor = ColorGreen
				case http.MethodPut, http.MethodPatch:
					methodColor = ColorYellow
				case http.MethodDelete:
					methodColor = ColorRed
				}

				reqID := middleware.GetReqID(r.Context())
				if reqID == "" {
					reqID = "-"
				}

				durationStr := formatDuration(duration)

				fmt.Printf(
					"%s[%s]%s %s%s%-6s%s %s %s%d%s %s%s%s %s(req_id: %s)%s\n",
					ColorGray, time.Now().Format("15:04:05"), ColorReset,
					ColorBold, methodColor, r.Method, ColorReset,
					r.URL.Path,
					ColorBold+statusColor, status, ColorReset,
					ColorMagenta, durationStr, ColorReset,
					ColorGray, reqID, ColorReset,
				)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
	return fmt.Sprintf("%.2fms", float64(d.Microseconds())/1000.0)
}

// PrintBanner prints a vibrant Pino-inspired startup banner in the terminal.
func PrintBanner(port, env, dbTarget string) {
	banner := fmt.Sprintf(`
%s%s┌─────────────────────────────────────────────────────────────┐
│  MINI ENTERPRISE GO BACKEND (Chi + Prisma + PostgreSQL)     │
└─────────────────────────────────────────────────────────────┘%s
  %s•%s %sServer URL:%s   http://localhost:%s
  %s•%s %sEnvironment:%s  %s
  %s•%s %sDatabase:%s     %s
  %s•%s %sFeatures:%s     JWT Auth • Pino Logger • Rate Limiter • PgBouncer
  %s•%s %sHealth:%s       http://localhost:%s/health
`,
		ColorBold, ColorCyan, ColorReset,
		ColorGreen, ColorReset, ColorBold, ColorReset, port,
		ColorGreen, ColorReset, ColorBold, ColorReset, env,
		ColorGreen, ColorReset, ColorBold, ColorReset, dbTarget,
		ColorGreen, ColorReset, ColorBold, ColorReset,
		ColorGreen, ColorReset, ColorBold, ColorReset, port,
	)
	fmt.Print(banner)
}
