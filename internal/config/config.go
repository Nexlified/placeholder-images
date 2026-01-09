package config

import (
	"flag"
	"os"
	"strconv"
)

const (
	DefaultSize               = 128
	DefaultBgColor            = "cccccc"
	DefaultFontColor          = "969696"
	DefaultAvatarBg           = "f0e9e9"
	DefaultAvatarFg           = "8b5d5d"
	DefaultAddr               = ":8080"
	DefaultDomain             = "localhost:8080"
	DefaultStaticDir          = "./static"
	CacheSize                 = 2000
	MinWidthForQuoteJoke      = 300 // Minimum width required to render quotes/jokes
	MinFontSize               = 16  // Minimum font size for readability
	MaxFontSize               = 48  // Maximum font size to avoid huge text
	MinTextLengthForSmallFont = 2   // Text longer than this uses smaller font (and may enable wrapping)
	// MinTextLengthForWrapping is kept for backward compatibility; prefer MinTextLengthForSmallFont.
	MinTextLengthForWrapping = MinTextLengthForSmallFont
	MinCharsPerLine          = 10 // Minimum characters per line for SVG text estimation
	// Rate limiting defaults
	DefaultRateLimitRPM  = 100 // Default requests per minute per IP
	DefaultRateLimitBurst = 10  // Default burst size for rate limiter
)

// ServerConfig represents runtime server settings.
type ServerConfig struct {
	Addr           string
	Domain         string
	StaticDir      string
	CacheSize      int
	RateLimitRPM   int // Requests per minute per IP
	RateLimitBurst int // Burst size for rate limiter
}

var (
	addrFlag           = flag.String("addr", "", "HTTP listen address (env ADDR)")
	domainFlag         = flag.String("domain", "", "Public domain for example URLs (env DOMAIN)")
	staticDirFlag      = flag.String("static-dir", "", "Directory for static files (env STATIC_DIR)")
	cacheSizeFlag      = flag.Int("cache-size", 0, "LRU cache size (env CACHE_SIZE)")
	rateLimitRPMFlag   = flag.Int("rate-limit-rpm", 0, "Rate limit requests per minute per IP (env RATE_LIMIT_RPM)")
	rateLimitBurstFlag = flag.Int("rate-limit-burst", 0, "Rate limit burst size (env RATE_LIMIT_BURST)")
)

// DefaultServerConfig returns sane defaults for local development.
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		Addr:           DefaultAddr,
		Domain:         DefaultDomain,
		StaticDir:      DefaultStaticDir,
		CacheSize:      CacheSize,
		RateLimitRPM:   DefaultRateLimitRPM,
		RateLimitBurst: DefaultRateLimitBurst,
	}
}

// LoadServerConfig reads defaults, then env, then flags.
func LoadServerConfig() ServerConfig {
	cfg := DefaultServerConfig()

	if addr := os.Getenv("ADDR"); addr != "" {
		cfg.Addr = addr
	}
	if domain := os.Getenv("DOMAIN"); domain != "" {
		cfg.Domain = domain
	}
	if staticDir := os.Getenv("STATIC_DIR"); staticDir != "" {
		cfg.StaticDir = staticDir
	}
	if cacheEnv := os.Getenv("CACHE_SIZE"); cacheEnv != "" {
		if n, err := strconv.Atoi(cacheEnv); err == nil && n > 0 {
			cfg.CacheSize = n
		}
	}
	if rateLimitRPMEnv := os.Getenv("RATE_LIMIT_RPM"); rateLimitRPMEnv != "" {
		if n, err := strconv.Atoi(rateLimitRPMEnv); err == nil && n > 0 {
			cfg.RateLimitRPM = n
		}
	}
	if rateLimitBurstEnv := os.Getenv("RATE_LIMIT_BURST"); rateLimitBurstEnv != "" {
		if n, err := strconv.Atoi(rateLimitBurstEnv); err == nil && n > 0 {
			cfg.RateLimitBurst = n
		}
	}

	if !flag.Parsed() {
		flag.Parse()
	}

	if addrFlag != nil && *addrFlag != "" {
		cfg.Addr = *addrFlag
	}
	if domainFlag != nil && *domainFlag != "" {
		cfg.Domain = *domainFlag
	}
	if staticDirFlag != nil && *staticDirFlag != "" {
		cfg.StaticDir = *staticDirFlag
	}
	if cacheSizeFlag != nil && *cacheSizeFlag > 0 {
		cfg.CacheSize = *cacheSizeFlag
	}
	if rateLimitRPMFlag != nil && *rateLimitRPMFlag > 0 {
		cfg.RateLimitRPM = *rateLimitRPMFlag
	}
	if rateLimitBurstFlag != nil && *rateLimitBurstFlag > 0 {
		cfg.RateLimitBurst = *rateLimitBurstFlag
	}

	return cfg
}
