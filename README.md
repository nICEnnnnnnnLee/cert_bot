# cert_bot
Obtain certs from Let's Encrypt.

Only support `dns-01` challenge.

Support `Cloudflare` API to deploy TXT record.

Support `Afraid` Cookie to deploy TXT record.

# Usage
```
Usage of cert_bot:
  -accountfile string
        the file that the account json data will be saved to/loaded from (will create new file if not exists) (default "account.json")
  -certfile string
        the file that the pem encoded certificate chain will be saved to (default "cert.pem")
  -contact string
        a list of comma separated contact emails to use when creating a new account (optional, dont include 'mailto:' prefix)
  -dirurl string
        acme directory url - defaults to lets encrypt v2 staging url if not provided.
         LetsEncryptProduction = https://acme-v02.api.letsencrypt.org/directory
         LetsEncryptStaging = https://acme-staging-v02.api.letsencrypt.org/directory
         ZeroSSLProduction = https://acme.zerossl.com/v2/DV90
         (default "https://acme-v02.api.letsencrypt.org/directory")
  -dns01file string
        the file that the dns01 json data will be loaded from (will exit if not exists) (default "dns01.json")
  -dnsserver string
        dnsServer to check txt record (default "8.8.8.8:53")
  -domains string
        a comma separated list of domains to issue a certificate for
  -exitifdns01fail
        exit if dns01 config is not valid, or just manualy set dns txt record (default true)
  -keyfile string
        the file that the pem encoded certificate private key will be saved to (default "privkey.pem")
  -txtmaxcheck int
        the max time trying to verify the txt record. program will continue after max retries no matter if the txt record is valid or not from local spec (default 30)
```

# Quick Start

+ Set `dns01.json`(Optional)  
If `dns01.json` do not valid and **-exitifdns01fail=false** pass to the cmd args, you can set the TXT record manually.
  <details>
      <summary>Cloudflare config</summary>


    
    If `api-email` is not empty, the auth uses header `X-Auth-Email` and `X-Auth-Key`; or else it just use header `Authorization`.  
    See <https://dash.cloudflare.com/profile/api-tokens>

    + `X-Auth-Email` + `X-Auth-Key`
    ```json
    {
      "type": "cloudflare",
      "config": {
        "api-email": "cf email account",
        "api-key": "global key",
        "domain": "the root domain in the dash borad panel. e.g. example.com"
      }
    }
    ```
    + `Authorization`
    ```json
    {
      "type": "cloudflare",
      "config": {
        "api-key": "dedicated token",
        "domain": "the root domain in the dash borad panel. e.g. example.com"
      }
    }
    ```
  </details>

  <details>
      <summary>Afraid config</summary>


    
    See <https://freedns.afraid.org/faq/#17>

    ```json
    {
      "type": "cloudflare",
      "config": {
        "dns_cookie": "get from browser",
        "domain_id": "number format id, get it from browser"
      }
    }
    ```
  </details>

+ Set `account.json`(Optional)  
The program will create `account.json` if it doesn't exist.  
  ```json
  {
    "privateKey":"-----BEGIN EC PRIVATE KEY-----\n...aaa\n...bbb\n-----END EC PRIVATE KEY-----\n",
    "url":"https://acme-staging-v02.api.letsencrypt.org/acme/acct/[0-9]+"
  }
  ```

+ Run `cet_bot`
  ```sh
  cet_bot -domains example.com,*.example.com
  ```

+ Get the outputs  
You will see `privkey.pem` and `cert.pem` in the same directory


# Specify input/output file path
See `Usage` for  help, or run help command
```sh
cet_bot -h
```

Here's an example.(replace `\` with `^` on windows)
```sh
mkdir -p ./certs/example.org
cet_bot -domains example.org,*.example.org              \
    -dns01file    ./certs/example.org/dns01.org.json    \
    -accountfile  ./certs/example.org/account.json      \
    -keyfile      ./certs/example.org/privkey.pem       \
    -certfile     ./certs/example.org/cert.pem          \
    -dnsserver 1.1.1.1:53
```

Or manually set the txt records:
```sh
mkdir -p ./certs/example.org
cet_bot -domains example.org,*.example.org              \
    -accountfile  ./certs/example.org/account.json      \
    -keyfile      ./certs/example.org/privkey.pem       \
    -certfile     ./certs/example.org/cert.pem          \
    -dnsserver 1.1.1.1:53 -exitifdns01fail=false
```
