# Kubernetes ConfigMap Reload

[![license](https://img.shields.io/github/license/jimmidyson/configmap-reload.svg?maxAge=2592000)](https://github.com/jimmidyson/configmap-reload)
[![Docker Stars](https://img.shields.io/docker/stars/jimmidyson/configmap-reload.svg?maxAge=2592000)](https://hub.docker.com/r/jimmidyson/comfigmap-reload/)
[![Docker Pulls](https://img.shields.io/docker/pulls/jimmidyson/configmap-reload.svg?maxAge=2592000)](https://hub.docker.com/r/jimmidyson/comfigmap-reload/)

**configmap-reload** is a simple binary to trigger a reload when a Kubernetes ConfigMap is updated.
It watches the mounted volume dir and notifies the target process that the config map has been changed.
It currently only supports sending an HTTP request, but in future it is expected to support sending OS
(e.g. SIGHUP) once Kubernetes supports pod PID namespaces.

It is available as a Docker image at https://hub.docker.com/r/jimmidyson/configmap-reload

### Usage

```
Usage of ./out/configmap-reload:
  -volume-dir string
        the config map volume directory to watch for updates
  -webhook-method string
        the HTTP method url to use to send the webhook (default "POST")
  -webhook-status-code int
        the HTTP status code indicating successful triggering of reload (default 200)
  -webhook-url string
        the url to send a request to when the specified config map volume directory has been updated<Paste>
```

### License

This project is [Apache Licensed](LICENSE.txt)

