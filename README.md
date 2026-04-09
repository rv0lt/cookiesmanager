# Traefik middleware to enable or disable the secure flag on sticky cookies

Based on: https://github.com/Lambda-IT/traefik-plugin-cookie-flags/tree/main

## Usage

Add the plugin in your static configuration

```yaml
# Static configuration
experimental:
  plugins:
    cookiesmanager:
      moduleName: github.com/rv0lt/cookiesmanager
      version: "0.1.0"
```

Use the plugin in your dynamic configuration like this

```yaml
# Dynamic configuration

http:
  routers:
    my-router:
      rule: host(`demo.localhost`)
      service: service-foo
      entryPoints:
        - websecure
      middlewares:
        - cookiesmanager

  services:
   service-foo:
      loadBalancer:
        servers:
          - url: http://127.0.0.1:5000
  
  middlewares:
    cookiesmanager:
      plugin:
        cookiesmanager:
          secure: true
```
