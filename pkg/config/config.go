package config

import (
	"github.com/joho/godotenv"
	"os"
	"path/filepath"
	"strconv"
)

func Load(serviceName string) {
	dir := "./configs"

	files := []string{
		filepath.Join(dir, ".env"),
		filepath.Join(dir, ".env.local"),
	}

	if serviceName != "" {
		files = append(files,
			filepath.Join(dir, ".env."+serviceName),
			filepath.Join(dir, ".env."+serviceName+".local"))
	}

	for _, file := range files {
		godotenv.Overload(file)
	}
}

func Get[T string | int | bool | float64](key string, zeroValue T) T {
	value := os.Getenv(key)
	if value == "" {
		return zeroValue
	}

	var result T
	switch any(zeroValue).(type) {
	case string:
		result = any(value).(T)
	case int:
		i, _ := strconv.Atoi(value)
		result = any(i).(T)
	case bool:
		b, _ := strconv.ParseBool(value)
		result = any(b).(T)
	case float64:
		f, _ := strconv.ParseFloat(value, 64)
		result = any(f).(T)
	}
	return result
}
