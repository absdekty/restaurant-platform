package config

import (
	"github.com/joho/godotenv"
	"os"
	"path/filepath"
	"strconv"
)

func Load(serviceName ...string) {
	dir := "./configs"

	files := make([]string, 0, 2+len(serviceName)*2)
	files = append(files,
		filepath.Join(dir, ".env"),
		filepath.Join(dir, ".env.local"))
	for _, name := range serviceName {
		if name == "" {
			continue
		}

		files = append(files,
			filepath.Join(dir, ".env."+name),
			filepath.Join(dir, ".env."+name+".local"))
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
