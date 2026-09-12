# Yandex Webmaster MCP

![Webmaster analytics connected to an AI agent through MCP](docs/assets/hero.png)

**Direct Webmaster access for AI agents.**

[![Tests](https://github.com/tenqz/yandex-webmaster-mcp/actions/workflows/tests.yml/badge.svg)](https://github.com/tenqz/yandex-webmaster-mcp/actions/workflows/tests.yml)
[![Code quality](https://github.com/tenqz/yandex-webmaster-mcp/actions/workflows/code-quality.yml/badge.svg)](https://github.com/tenqz/yandex-webmaster-mcp/actions/workflows/code-quality.yml)
[![MIT](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

A small, self-hosted Go server that lets MCP agents read Yandex Webmaster hosts, search queries, indexing, sitemaps, diagnostics and recrawl quota. Streamable HTTP, one OAuth token per installation. Independent community project; not affiliated with Yandex.

Ask your agent to compare SQI, list popular queries, explain indexing samples or check recrawl quota. The server retrieves the data; your agent interprets it. `submit_recrawl` is the only write tool and consumes daily quota.

## Try it without credentials

Requires Docker with Compose. This builds the current checkout and uses clearly labelled synthetic data, with no Yandex requests:

```bash
git clone https://github.com/tenqz/yandex-webmaster-mcp.git
cd yandex-webmaster-mcp
docker compose -f compose.demo.yml up -d --build
docker compose -f compose.demo.yml exec mcp /mcp-server --smoke http://127.0.0.1:8080/mcp
```

Connect your MCP client to `http://localhost:8080/mcp` with header `Authorization: Bearer demo-token`. The smoke command checks discovery and a representative set of tools. Stop with `docker compose -f compose.demo.yml down` before starting a live installation.

## Connect Yandex Webmaster

1. Follow [Yandex setup](docs/yandex-setup.md): register an OAuth application with `webmaster:hostinfo` (and `webmaster:verify` if you need verification later) and obtain a token. Paste the token into `.env` as `YANDEX_WEBMASTER_TOKEN`. There is no in-server browser login flow.
2. Copy `.env.example` to `.env`. Replace `MCP_AUTH_TOKEN` with the output of `openssl rand -hex 32`.
3. Start the server:

```bash
docker compose up -d --build
curl --fail http://localhost:8080/health
docker compose exec mcp /mcp-server --smoke http://127.0.0.1:8080/mcp
```

The smoke check uses the first accessible host. Yandex access and quota errors are reported by the tools.

For access from another machine, use [HTTPS deployment](docs/deployment.md). Connect agents using the templates in [MCP clients and tools](docs/mcp.md). A health response confirms the process is running; the smoke check also verifies MCP and Yandex access.

## Tools

Call `list_hosts` first and pass `host_id` exactly (`https:example.com:443`). Most tools are read-only. `submit_recrawl` writes.

| Tool | Result |
| --- | --- |
| `get_user` | OAuth user id |
| `list_hosts` | Sites and verification |
| `get_host` | Site details |
| `get_summary` | SQI, pages, problems |
| `get_sqi_history` | SQI over time |
| `get_diagnostics` | Site diagnostics |
| `get_popular_queries` | Queries with shows/clicks/position |
| `get_query_history` | Aggregated query history |
| `get_single_query_history` | History for one `query_id` |
| `get_query_analytics` | Query↔URL report |
| `get_indexing_history` / `get_indexing_samples` | Robot downloads |
| `get_insearch_history` / `get_insearch_samples` | Pages in search |
| `get_search_events_history` / `get_search_events_samples` | Added/removed pages |
| `get_external_links` / `get_external_links_history` | Backlinks |
| `get_broken_internal_links` / `get_broken_internal_links_history` | Broken internal links |
| `get_sitemaps` / `get_sitemap` | Detected sitemaps |
| `get_user_sitemaps` / `get_user_sitemap` | User-added sitemaps |
| `get_important_urls` / `get_important_url_history` | Important pages |
| `get_recrawl_quota` / `get_recrawl_queue` / `get_recrawl_task` | Recrawl quota and tasks |
| `submit_recrawl` | Queue a URL (write) |
| `get_feeds` / `get_feed_status` | Feeds |
| `get_region_ids` / `get_feed_regions` | Static region catalogs |

Each tool publishes an output schema. Failures include a stable error code, a safe message and retry information when available.

## Configuration

| Variable | Default | Purpose |
| --- | --- | --- |
| `HTTP_ADDR` | `:8080` | Listening address; Compose exposes it on host loopback |
| `MCP_PATH` | `/mcp` | MCP endpoint |
| `MCP_AUTH_TOKEN` | Required | Shared bearer secret |
| `YANDEX_WEBMASTER_TOKEN` | Required unless demo | Yandex OAuth token |
| `YANDEX_WEBMASTER_BASE_URL` | `https://api.webmaster.yandex.net/v4` | API root |
| `YANDEX_REQUEST_TIMEOUT` | `30s` | Total Yandex request budget, including queue and retries |
| `YANDEX_MAX_CONCURRENT` | `8` | Concurrent Yandex work per process |
| `YANDEX_MAX_ATTEMPTS` | `3` | Total attempts for transient failures |
| `MCP_MAX_BODY_BYTES` | `1048576` | Maximum incoming MCP request size |
| `MCP_DEMO` | `false` | Synthetic data; cannot be combined with a Yandex token |
| `MCP_ALLOW_INSECURE` | `false` | Token-free debugging; requires a loopback listening address |

Yandex responses are capped at 8 MiB. Logs contain operation outcomes and durations, without request arguments, tokens or raw Yandex error bodies.

## Troubleshooting

- **401:** check the MCP bearer header and the server's token.
- **Yandex 401/403:** recreate the OAuth token and confirm Webmaster scopes and host access. Use `host_id` exactly as returned by `list_hosts`.
- **429:** respect retry information and reduce concurrency. Retries share the request time budget.
- **413:** reduce the MCP request size; raise the configured bound only when necessary.
- **Timeout:** reduce query scope or concurrency before raising `YANDEX_REQUEST_TIMEOUT`.
- **Empty analytics:** verify the host, dates and that the site has Webmaster data. Empty rows can be a valid response.

## Development and releases

Use Go 1.25+ with a current security patch; release containers use Go 1.26.6. Run `make check` for formatting, module consistency, vet, build, race tests and pinned lint. Run `make docker-test` for a container smoke check. See [testing](docs/testing.md), [architecture](docs/architecture.md), [contributing](CONTRIBUTING.md) and [security](SECURITY.md).

The release workflow verifies a `v1.0.0` tag against the binary version, checks both container architectures and publishes versioned images to `ghcr.io/tenqz/yandex-webmaster-mcp`. Images become available after the tag workflow succeeds. See [deployment](docs/deployment.md) for updates and rollback. Licensed under [MIT](LICENSE).

## Contact

Website: [opatsay.com](https://opatsay.com/)
