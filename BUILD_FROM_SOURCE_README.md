# Build Instructions

Based on upstream tag `v0.7.1`

---
## Build

```
export DOCKER_IMAGE_NAME="configmap-reload"
export DOCKER_IMAGE_TAG="0.7.1"
./build-image.sh 
```

## Push to OCIR

`docker image push <docker-image-repo>/configmap-reload:0.7.1`
