package models

type contextKey string

const (
	TraceIDKey contextKey = "trace_id"
	LoggerKey  contextKey = "logger"
)
