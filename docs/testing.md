# Testing

Run `make check` with Go 1.25+ and a current security patch. It verifies formatting, `go mod tidy -diff`, vet, build, race tests and the same pinned linter used by CI. The linter is installed under `.tools/`; the global installation is untouched. An ignored vendor directory is not used.

`make docker-test` builds the runtime and calls a representative set of tools against deterministic demo data. `docker compose -f compose.demo.yml up -d --build` reproduces the documented onboarding path. CI runs Go 1.25 and 1.26 and smoke-tests Linux amd64 and arm64 images. The security workflow runs govulncheck on changes and weekly.

The HTTP integration tests use the real MCP SDK, bearer middleware and Webmaster client with a fake upstream transport. They cover successful calls, validation, upstream failures and OAuth headers without real credentials.

For live acceptance, configure your own Yandex token and HTTPS endpoint, then run `MCP_AUTH_TOKEN=... bin/mcp-server --smoke https://your-host/mcp`. Demo tests do not prove Yandex permissions, DNS/TLS configuration or a particular desktop client's behavior. Do not supply real credentials to PR workflows.

ACDD requires exactly one file per commit. Verify source changes first, then add regression tests in a separate passing commit.
