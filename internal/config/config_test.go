package config_test

import (
	"testing"

	"github.com/tenqz/yandex-webmaster-mcp/internal/config"
)

// TestLoadRequiresAuthToken documents that a public MCP endpoint must not start without a shared secret.
func TestLoadRequiresAuthToken(t *testing.T) {
	t.Setenv("MCP_AUTH_TOKEN", "")
	t.Setenv("MCP_ALLOW_INSECURE", "")
	t.Setenv("YANDEX_WEBMASTER_TOKEN", "oauth-token")

	_, err := config.Load()

	if err == nil {
		t.Fatal("expected an error when MCP_AUTH_TOKEN is empty")
	}
}

// TestLoadAllowsInsecureWithoutToken documents the local-debug escape hatch.
func TestLoadAllowsInsecureWithoutToken(t *testing.T) {
	t.Setenv("MCP_AUTH_TOKEN", "")
	t.Setenv("MCP_ALLOW_INSECURE", "true")
	t.Setenv("YANDEX_WEBMASTER_TOKEN", "oauth-token")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !cfg.AllowInsecure {
		t.Fatal("expected AllowInsecure to be true")
	}
}

// TestLoadRequiresYandexToken documents that Yandex OAuth is mandatory outside demo mode.
func TestLoadRequiresYandexToken(t *testing.T) {
	t.Setenv("MCP_AUTH_TOKEN", "secret")
	t.Setenv("YANDEX_WEBMASTER_TOKEN", "")
	t.Setenv("MCP_DEMO", "")

	_, err := config.Load()

	if err == nil {
		t.Fatal("expected an error when YANDEX_WEBMASTER_TOKEN is missing")
	}
}

// TestLoadRejectsUnsafeAndInvalidSettings protects public binds, route patterns and resource budgets.
func TestLoadRejectsUnsafeAndInvalidSettings(t *testing.T) {
	for _, tc := range []struct{ name, value string }{
		{"HTTP_ADDR", ":8080"},
		{"MCP_PATH", "/health"},
		{"MCP_PATH", "/{wildcard}"},
		{"YANDEX_REQUEST_TIMEOUT", "0s"},
		{"YANDEX_MAX_CONCURRENT", "0"},
		{"YANDEX_MAX_ATTEMPTS", "20"},
		{"MCP_MAX_BODY_BYTES", "0"},
	} {
		t.Run(tc.name+tc.value, func(t *testing.T) {
			t.Setenv("MCP_AUTH_TOKEN", "")
			t.Setenv("MCP_ALLOW_INSECURE", "true")
			t.Setenv("YANDEX_WEBMASTER_TOKEN", "oauth-token")
			t.Setenv(tc.name, tc.value)
			if _, err := config.Load(); err == nil {
				t.Fatal("unsafe settings accepted")
			}
		})
	}
}

// TestDemoConfiguration prevents accidentally displaying fixtures as a configured real account.
func TestDemoConfiguration(t *testing.T) {
	t.Setenv("MCP_DEMO", "true")
	t.Setenv("MCP_AUTH_TOKEN", "test")
	t.Setenv("YANDEX_WEBMASTER_TOKEN", "")
	cfg, err := config.Load()
	if err != nil || !cfg.Demo {
		t.Fatalf("demo: %v", err)
	}
	t.Setenv("YANDEX_WEBMASTER_TOKEN", "oauth-token")
	if _, err := config.Load(); err == nil {
		t.Fatal("demo accepted a live OAuth token")
	}
}
