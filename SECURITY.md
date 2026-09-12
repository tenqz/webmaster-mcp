# Security policy

Security fixes target the latest 1.0.x release. Use current patches and rebuild containers when Go or dependency advisories require it.

Report suspected vulnerabilities privately to smmartbiz@gmail.com with reproduction steps and affected versions. Do not include OAuth tokens, bearer tokens or private Webmaster data. Please allow investigation before public disclosure.

The MCP bearer token grants access to every host visible to the installation's Yandex OAuth token. Use HTTPS for remote connections and a separate installation for each trust boundary. `/health` is intentionally unauthenticated and contains no Webmaster data. Insecure mode is for loopback debugging only. Demo data is synthetic; demo mode still requires authentication by default.

Keep tokens outside version control and container images, restrict file permissions and rotate exposed secrets. Automated tests use fake Webmaster responses and require no real credentials.
