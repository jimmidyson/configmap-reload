# Build Instructions

Based on branch `v0.3.0`

---
## Build

```
export DOCKER_IMAGE_NAME="configmap-reloader"
export DOCKER_IMAGE_TAG="0.3"
./build-image.sh 
```

## Push to OCIR

`docker image push <docker-image-repo>/configmap-reloader:0.3`