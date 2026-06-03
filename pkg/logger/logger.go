package logger

import (
	"log"
	"os"
)

var (
	Info  *log.Logger
	Warn  *log.Logger
	Error *log.Logger
)

func Init(serviceName string) {
	prefix := "[" + serviceName + "] "

	Info = log.New(os.Stdout, prefix+"[INFO] ", log.LstdFlags)
	Warn = log.New(os.Stdout, prefix+"[WARN] ", log.LstdFlags)
	Error = log.New(os.Stderr, prefix+"[ERROR] ", log.LstdFlags|log.Lshortfile)
}
