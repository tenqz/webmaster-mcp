# Yandex Webmaster access

The MCP server authenticates to Yandex with a **user OAuth token**. The token owner must already have the sites in Webmaster. This installation does not open a browser or run an OAuth redirect.

## 1. OAuth application

1. Open [Yandex OAuth app registration](https://oauth.yandex.ru/client/new).
2. Create an application (web services). Redirect URI can be `https://oauth.yandex.ru/verification_code`.
3. Enable `webmaster:hostinfo`. Enable `webmaster:verify` only if you later need verification tools.
4. Copy the ClientID.

## 2. Token

Open this URL, replacing `CLIENT_ID`:

```text
https://oauth.yandex.ru/authorize?response_type=token&client_id=CLIENT_ID
```

Authorize in the browser and copy the access token from the resulting page. Tokens expire (commonly after several months). Put the token in `.env` as `YANDEX_WEBMASTER_TOKEN`. Do not commit this file.

The MCP server itself only reads `YANDEX_WEBMASTER_TOKEN`. It does not exchange a client secret.

## 3. Verify

After deploy, ask the agent to call `list_hosts`. You should see the sites available to that Yandex account. If the list is empty, the token cannot see those hosts.
