package config

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Config holds process-wide settings loaded from the environment.
// It is the only place that reads OS environment variables.
type Config struct {
	ReadOnly       bool
	Demo           bool
	RequestTimeout time.Duration
	MaxConcurrent  int
	MaxAttempts    int
	MaxBodyBytes   int64
	// Addr is the TCP host:port the HTTP server binds to.
	Addr string
	// MCPPath is the HTTP path that serves the MCP Streamable transport.
	MCPPath string
	// AuthToken is the shared bearer secret agents must send.
	AuthToken string
	// AllowInsecure disables bearer auth. Use only for local debugging.
	AllowInsecure bool
	// YandexToken is the OAuth token for api.webmaster.yandex.net.
	YandexToken string
	// YandexBaseURL is the Webmaster API root, including the /v4 prefix.
	YandexBaseURL string
}

// Load reads configuration from environment variables and applies defaults.
// It returns an error when a production-unsafe combination is detected.
func Load() (Config, error) {
	authToken, err := Secret("MCP_AUTH_TOKEN")
	if err != nil {
		return Config{}, err
	}
	yandexToken, err := Secret("YANDEX_WEBMASTER_TOKEN")
	if err != nil {
		return Config{}, err
	}
	cfg := Config{
		ReadOnly:      truthy(os.Getenv("MCP_READ_ONLY")),
		Demo:          truthy(os.Getenv("MCP_DEMO")),
		Addr:          envOr("HTTP_ADDR", ":8080"),
		MCPPath:       envOr("MCP_PATH", "/mcp"),
		AuthToken:     authToken,
		AllowInsecure: truthy(os.Getenv("MCP_ALLOW_INSECURE")),
		YandexToken:   yandexToken,
		YandexBaseURL: envOr("YANDEX_WEBMASTER_BASE_URL", "https://api.webmaster.yandex.net/v4"),
	}

	if cfg.AllowInsecure && strings.TrimSpace(os.Getenv("HTTP_ADDR")) == "" {
		cfg.Addr = "127.0.0.1:8080"
	}
	host, _, err := net.SplitHostPort(cfg.Addr)
	if err != nil {
		return Config{}, fmt.Errorf("HTTP_ADDR must be host:port")
	}
	if cfg.AllowInsecure {
		ip := net.ParseIP(host)
		if host != "localhost" && (ip == nil || !ip.IsLoopback()) {
			return Config{}, fmt.Errorf("MCP_ALLOW_INSECURE requires a loopback HTTP_ADDR")
		}
	}
	cfg.RequestTimeout, err = time.ParseDuration(envOr("YANDEX_REQUEST_TIMEOUT", "30s"))
	if err != nil || cfg.RequestTimeout < time.Second || cfg.RequestTimeout > 5*time.Minute {
		return Config{}, fmt.Errorf("YANDEX_REQUEST_TIMEOUT must be between 1s and 5m")
	}
	cfg.MaxConcurrent, err = boundedInt("YANDEX_MAX_CONCURRENT", 8, 1, 128)
	if err != nil {
		return Config{}, err
	}
	cfg.MaxAttempts, err = boundedInt("YANDEX_MAX_ATTEMPTS", 3, 1, 5)
	if err != nil {
		return Config{}, err
	}
	bodyLimit, err := boundedInt("MCP_MAX_BODY_BYTES", 1<<20, 1024, 16<<20)
	if err != nil {
		return Config{}, err
	}
	cfg.MaxBodyBytes = int64(bodyLimit)
	if !regexp.MustCompile(`^/[a-zA-Z0-9_-]+(?:/[a-zA-Z0-9_-]+)*$`).MatchString(cfg.MCPPath) || cfg.MCPPath == "/health" {
		return Config{}, fmt.Errorf("MCP_PATH must be a literal path distinct from /health")
	}
	if !cfg.AllowInsecure && (strings.TrimSpace(cfg.AuthToken) == "" || cfg.AuthToken == "replace-me-with-a-long-random-token") {
		return Config{}, fmt.Errorf("MCP_AUTH_TOKEN is required unless MCP_ALLOW_INSECURE=true")
	}
	if !cfg.Demo && cfg.YandexToken == "" {
		return Config{}, fmt.Errorf("set YANDEX_WEBMASTER_TOKEN")
	}
	if cfg.Demo && cfg.YandexToken != "" {
		return Config{}, fmt.Errorf("MCP_DEMO must not be combined with YANDEX_WEBMASTER_TOKEN")
	}
	if strings.TrimRight(cfg.YandexBaseURL, "/") == "" {
		return Config{}, fmt.Errorf("YANDEX_WEBMASTER_BASE_URL must not be empty")
	}
	cfg.YandexBaseURL = strings.TrimRight(cfg.YandexBaseURL, "/")
	return cfg, nil
}

// envOr returns the environment value or fallback when the variable is empty.
func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

// truthy reports whether value looks like an enabled boolean flag.
func truthy(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func boundedInt(name string, fallback, minValue, maxValue int) (int, error) {
	value, err := strconv.Atoi(envOr(name, strconv.Itoa(fallback)))
	if err != nil || value < minValue || value > maxValue {
		return 0, fmt.Errorf("%s must be between %d and %d", name, minValue, maxValue)
	}
	return value, nil
}

// Secret reads an environment secret or a mounted secret file, never both.
func Secret(name string) (string, error) {
	value, path := os.Getenv(name), os.Getenv(name+"_FILE")
	if value != "" && path != "" {
		return "", fmt.Errorf("set only %s or %s_FILE", name, name)
	}
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("cannot read %s_FILE", name)
		}
		value = string(data)
	}
	return strings.TrimSpace(value), nil
}
