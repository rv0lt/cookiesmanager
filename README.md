# Traefik middleware cookies manager plugin

Simple plugin that ensures cookies of requests contains or does not contains some values. If the specified cookie is not in the request, then the plugin add the cookie.

Based on: https://github.com/theoguidoux/cookiesmanager

## TODO: Explain changes from main project

## Demo

Navigate to the [demo](https://github.com/rv0lt/cookiesmanager/tree/main/demo) 
folder in the repo to run a quick docker-compose
demonstration of this plugin. It includes additional information on
how to start and use the demo.

The demo also illustrates how the plugin can be loaded in a traefik
docker image and used without relying on the traefik Pilot. For more
information about packaging plugins in an image see 
[here](https://traefik.io/blog/using-private-plugins-in-traefik-proxy-2-5/).

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
          adder:
            - name: cookie1
              value: "foo=bar"
            - name: "SameSite"
              value: "none"
          remover:
            - name: "cookie2"
              value: "foo=bar"
```

The middleware defined above would make it so that requests to `service-foo` 
have `cookie1` containing `foo=bar`, `SameSite` containing `none` and `cookie2` not containing `foo=bar`.
