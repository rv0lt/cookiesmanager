# Traefik middleware cookies manager plugin

TODO

Based on: https://github.com/Lambda-IT/traefik-plugin-cookie-flags/tree/main

## Demo

## Usage

Add the plugin in your static configuration

```yaml
# Static configuration
experimental:
  plugins:
    cookiesmanager:
      moduleName: github.com/rv0lt/cookiesmanager
      version: "0.0.1"
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
        - web
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
