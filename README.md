# Kubernetes KeyValueStore TLS Reloader

**kvs-tls-reload** is a simple binary to trigger a reload of a Redis compatible KeyValueStore when
Kubernetes TLS Secrets, mounted into pods, are updated.

It watches mounted secret volume dirs for updated certificate files. After an update, it connects to
the KeyValueStore and reloads the certificates (by CONFIG SET command) without restarting the service
or pod. Therefore, the supplied user account needs to have permission to issue these commands.

The script is supposed to run in a sidecar container to be able to access the pod's filesystem and
network.

The Docker image is available from ghcr.io at <https://github.com/ninech/kvs-tls-reloader/pkgs/container/kvs-tls-reloader>.

### Usage

```
Usage: kvs-tls-reload --cert-dir=STRING [flags]

Reloads a KeyValueStore's TLS cert and key when they get replaced in the filesystem.

Flags:
  -h, --help                             Show context-sensitive help.
      --cert-dir=STRING                  The certificate directory to watch for updates ($KVS_CERT_DIR).
      --web.listen-address=":9533"       Address to listen on for web interface and telemetry.
      --web.telemetry-path="/metrics"    Path under which to expose metrics.
      --kvs-host="127.0.0.1"             Host where the KeyValueStore is running ($KVS_HOST).
      --kvs-port=6379                    The port the KeyValueStore is listening on ($KVS_PORT).
      --kvs-tls-enabled                  Connect to the KeyValueStore using TLS ($KVS_TLS_ENABLED).
      --kvs-user="default"               User for the KeyValueStore ($KVS_USER).
      --kvs-password=""                  Password for the KeyValueStore ($KVS_PASSWORD).
      --cert-filename="tls.crt"          Filename of the tls cert ($KVS_CERT_FILENAME).
      --key-filename="tls.key"           Filename of the tls key ($KVS_KEY_FILENAME).
      --ca-filename="ca.crt"             Filename of the ca cert ($KVS_CA_FILENAME).

```

### License

This project is [Apache Licensed](LICENSE.txt)

