# cert_bot
Obtain certs from Let's Encrypt.

Only support `dns-01` challenge.

Support `Cloudflare` API to deploy TXT record.

Support `Afraid` Cookie to deploy TXT record.

Support `He.net` DDNS API to deploy TXT record.

# Mode
## http server
Set environment `Mode` to `server`.

You can edit environment values refering to codes below:
```
UrlPrefix             = GetEnvOr("UrlPrefix", "/xx")            // {UrlPrefix}/api  {UrlPrefix}/static 
bindAddr              = GetEnvOr("BindAddr", "127.0.0.1:8080")
proxyURL              = GetEnvOr("ProxyUrl", "")                // the app will use ProxyUrl for http request
certPath              = GetEnvOr("CertPath", "")                // if CertPath and KeyPath not empty, server is HTTPS, else HTTP
keyPath               = GetEnvOr("KeyPath", "")
```

### Github OAuth
You can use Github OAuth to protect secrets.

Details of configs are as follows.
```
oauthCookieFormat     = GetEnvOr("OAuthCookieFormat", `%s=%s; domain=%s; path=%s; max-age=%s; secure; HttpOnly; SameSite=Lax`)
oauthCookieNamePrefix = GetEnvOr("OAuthCookieNamePrefix", "crtbot")
oauthCookiePath       = GetEnvOr("OAuthCookiePath", UrlPrefix)
oauthCookieTTL        = GetEnvOr("OAuthCookieTTL", "3600")
oauthCookieTTLInt64   int64
oauthClientId         = GetEnvOr("OAuthClientId", "")
OAuthClientSecret     = GetEnvOr("OAuthClientSecret", "")
oauthValidUsers       = GetEnvOr("OAuthValidUsers", "")
```

The important configs are **OAuthClientId**, **OAuthClientSecret** and **OAuthValidUsers**.

+ Prepare a domain `example.com`
+ Homepage URL: `https://example.com{UrlPrefix}/static/`
+ Callback URL: `https://example.com{UrlPrefix}/api/oauth`
+ Follow the [document](https://docs.github.com/en/developers/apps/building-oauth-apps/creating-an-oauth-app) to get **OAuthClientId** and **OAuthClientSecret**
+ **OAuthValidUsers**: Your github login account, not email or phone number. For multi-users, use `,` to seperate. Case in-sensetive


### Nginx config example

```
server {
    listen [::]:443 ssl http2;
    listen  443 ssl http2;
    ssl_certificate       /path/to/pem;
    ssl_certificate_key   /path/to/key;
    server_name           example.com;
    
    ssl_protocols         TLSv1 TLSv1.1 TLSv1.2;
    ssl_ciphers           HIGH:!aNULL:!MD5;
    root html;
    error_page 404    /404.html;
    
    #PROXY-START/
    location {UrlPrefix} {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $http_host;
        proxy_request_buffering off;
    }
    #PROXY-END/
}
```


## cli
Set environment `Mode` to `cli`, then see the [doc for cli](/README_CLI.md)

# Quick Start
