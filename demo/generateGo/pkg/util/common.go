package util

import (
	"os"
	"strconv"
	"strings"
	"time"
)

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return strings.TrimSpace(value)
	}
	return fallback
}

func BoolEnv(key string, fallback bool) bool {
	return transEnv(key, fallback).(bool)
}

func IntEnv(key string, fallback int) int {
	return transEnv(key, fallback).(int)
}

func DurationEnv(key string, fallback time.Duration) time.Duration {
	return transEnv(key, fallback).(time.Duration)
}

func transEnv(key string, fallback any) any {
	value, ok := os.LookupEnv(key)
	value = strings.TrimSpace(value)

	if !ok || value == "" {
		return fallback
	}

	var parsed any
	var err error

	switch fallback.(type) {
	case int:
		parsed, err = strconv.Atoi(value)
	case bool:
		parsed, err = strconv.ParseBool(value)
	case time.Duration:
		parsed, err = time.ParseDuration(value)
	}

	if err != nil {
		return fallback
	}
	return parsed
}

func SplitSCV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func IsDev(env string) bool {
	switch strings.ToLower(strings.TrimSpace(env)) {
	case "", "dev", "develop", "development", "local":
		return true
	default:
		return false
	}
}
