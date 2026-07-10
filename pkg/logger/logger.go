package logger

import (
	"log/slog"
	"os"
)

func SetupLogger(env, service string) {
	var handler slog.Handler

	switch env {
	case "prod":
		handler = slog.NewJSONHandler(os.Stdout,
			&slog.HandlerOptions{
				Level: slog.LevelInfo,
			})
	case "dev":
		handler = slog.NewTextHandler(os.Stdout,
			&slog.HandlerOptions{
				Level:     slog.LevelDebug,
				AddSource: true,
			})
	default:
		handler = slog.NewTextHandler(os.Stdout,
			&slog.HandlerOptions{
				Level:     slog.LevelDebug,
				AddSource: true,
			})
	}

	logger := slog.New(handler)
	logger = logger.With("service", service)

	slog.SetDefault(logger)
}
