# Deployment

The installation shares one Yandex OAuth token and MCP bearer token across a trusted owner or team. Anyone with the bearer can call every tool against every host visible to that OAuth token. Use separate installations for separate trust boundaries.

## HTTPS

Configure the live installation from the README first. Point a public DNS name to the server and allow inbound TCP 80/443 (UDP 443 is optional). Caddy obtains and renews certificates:

```bash
export MCP_HOST=webmaster.example.com
docker compose -f docker-compose.yml -f compose.https.yml up -d --build
```

Connect agents to `https://webmaster.example.com/mcp` with the bearer header. Keep the token private. The proxy forwards to the container network; port 8080 is bound only to host loopback. Do not expose plaintext HTTP outside the trusted local machine. Keep Caddy's data volume across restarts for certificate renewal.

To verify the external route with a locally built binary:

```bash
make build
MCP_AUTH_TOKEN='your-private-token' bin/mcp-server --smoke https://webmaster.example.com/mcp
```

This checks real access when a Yandex token is configured. Do not paste real tokens or responses into public issues.

## Versioned image

After a release workflow publishes the image, replace source builds with:

```bash
export MCP_VERSION=1.0.0
docker compose pull
docker compose up -d --no-build
```

For the HTTPS deployment, include both `-f` options on these commands. Pin an explicit version or image digest. To roll back, set `MCP_VERSION` to a previously published version, pull and recreate the service. Configuration and credentials stay outside the image. The server stores no analytics database.

The runtime is a non-root static binary with CA certificates, a read-only filesystem, dropped Linux capabilities, and bounded process/memory limits in Compose. Rotate the MCP bearer by updating `.env` and recreating the service. Rotate the Yandex OAuth token in Yandex OAuth and replace the env value.
