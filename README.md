# Kubernetes ConfigMap Reload

[![license](https://img.shields.io/github/license/fitsegreat/configmap-reload.svg?maxAge=2592000)](https://github.com/fitsegreat/configmap-reload)
[![Docker Stars](https://img.shields.io/docker/stars/fitsegreat/configmap-reload.svg?maxAge=2592000)](https://ghcr.io/v2/fitsegreat/configmap-reload/)
[![Docker Pulls](https://img.shields.io/docker/pulls/fitsegreat/configmap-reload.svg?maxAge=2592000)](https://ghcr.io/v2/fitsegreat/configmap-reload/)

**configmap-reload** is a simple binary to trigger a reload when Kubernetes ConfigMaps or Secrets, mounted into pods,
are updated.
It watches mounted volume dirs and notifies the target process that the config map has been changed.

### Usage

```
Usage of ./out/configmap-reload:
  --volume-dir value
        the config map volume directory to watch for updates; may be used multiple times
  --web.listen-address string
    	  address to listen on for web interface and telemetry. (default ":9533")
  --web.telemetry-path string
    	  path under which to expose metrics. (default "/metrics")
  --webhook-method string
        the HTTP method url to use to send the webhook (default "POST")
  --webhook-status-code int
        the HTTP status code indicating successful triggering of reload (default 200)
  --webhook-url string
        the url to send a request to when the specified config map volume directory has been updated (ignored if --target-process-name is set)
  --webhook-retries integer
        the amount of times to retry the webhook reload request
  --target-process-name
        Name of the target application process to send SIGHUP to when a filesystem change is detected; may be used multiple times. If empty, a webhook will be triggered instead.
```

### License

This project is [Apache Licensed](LICENSE.txt)

