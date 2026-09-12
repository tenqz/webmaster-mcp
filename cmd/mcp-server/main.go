package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/tenqz/yandex-webmaster-mcp/internal/buildinfo"
	"github.com/tenqz/yandex-webmaster-mcp/internal/config"
	"github.com/tenqz/yandex-webmaster-mcp/internal/httpserver"
	"github.com/tenqz/yandex-webmaster-mcp/internal/mcpserver"
	"github.com/tenqz/yandex-webmaster-mcp/internal/webmaster"
)

func main() {
	version := flag.Bool("version", false, "Print version and exit")
	healthURL := flag.String("healthcheck", "", "Check an HTTP health URL and exit")
	smokeURL := flag.String("smoke", "", "Verify MCP tools at a URL; requires MCP_AUTH_TOKEN and Yandex access or demo mode")
	flag.Parse()
	if *version {
		fmt.Println(buildinfo.Version)
		return
	}
	if *healthURL != "" {
		if err := checkHealth(*healthURL); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if *smokeURL != "" {
		token, err := config.Secret("MCP_AUTH_TOKEN")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := smoke(*smokeURL, token); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("invalid configuration", "err", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	api, err := newAPI(cfg)
	if err != nil {
		slog.Error("yandex webmaster client", "err", err)
		os.Exit(1)
	}

	mcpServer := mcpserver.New(api, cfg.ReadOnly)
	mcpHandler := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server {
		return mcpServer
	}, &mcp.StreamableHTTPOptions{Stateless: true})

	handler := httpserver.New(cfg, mcpHandler)
	httpServer := httpserver.NewServer(cfg.Addr, handler)
	httpServer.WriteTimeout = cfg.RequestTimeout + 15*time.Second
	httpServer.BaseContext = func(net.Listener) context.Context { return ctx }

	go func() {
		slog.Info("mcp server listening",
			"addr", cfg.Addr,
			"mcpPath", cfg.MCPPath,
			"insecure", cfg.AllowInsecure,
			"demo", cfg.Demo,
			"version", mcpserver.ServerVersion,
		)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("http server", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("http shutdown", "err", err)
		os.Exit(1)
	}
}

// newAPI builds the live Webmaster client from the configured OAuth token.
func newAPI(cfg config.Config) (webmaster.Webmaster, error) {
	if cfg.Demo {
		return webmaster.NewDemo(), nil
	}
	return webmaster.NewClient(http.DefaultClient, cfg.YandexToken, cfg.YandexBaseURL, webmaster.Options{
		RequestTimeout: cfg.RequestTimeout,
		MaxConcurrent:  cfg.MaxConcurrent,
		MaxAttempts:    cfg.MaxAttempts,
	}), nil
}
