package utils

import (
	"crypto/md5"
	"fmt"
	"strconv"
)

// ParseIntOrDefault converts the string to int or returns the default.
func ParseIntOrDefault(s string, def int) int {
	if s == "" {
		return def
	}
	i, err := strconv.Atoi(s)
	if err != nil || i <= 0 {
		return def
	}
	return i
}

// GenerateColorHash returns deterministic hex derived from input.
func GenerateColorHash(seed string) string {
	h := md5.Sum([]byte(seed))
	return fmt.Sprintf("%02x%02x%02x", h[0], h[1], h[2])
}
