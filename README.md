# oci-images-from-k8s-yaml

## What

- Send it a k8s manifest file, and you get a (newline separated) list of the OCI images those pods would run.

## Why

- Sometimes you want to do things for the list of images in a given set of kubernetes manifests
- ...like scanning them for vulnerabilities.

## Alternatives considered

1. If one can pick Syft/Grype, this looks like it'll solve the same problem:
   - https://github.com/anchore/syft/issues/2729
   - https://github.com/anchore/grype/issues/1259
   - https://github.com/anchore/syft/issues/562
1. But if one must use another image scanning tool (🙉), building this myself is the best I've found (yet?).

## Usage

### Get images for a manifest:

```shell
$ go run main.go examples/statefulset.yaml
registry.k8s.io/nginx-slim:0.8
gcr.io/google-containers/sidecar
kiwigrid/k8s-sidecar
```

### Get images for all manifests in a folder:

```shell
$ go run main.go examples/* | sort -u
busybox:1.28
gcr.io/google-containers/busybox
gcr.io/google-containers/sidecar
kiwigrid/k8s-sidecar
nginx
perl
registry.k8s.io/nginx-slim:0.8
```

### Scan images from a manifest:

```shell
# Syft:
$ go run main.go examples/job.yaml | xargs syft

# Snyk:
$ go run main.go examples/job.yaml | xargs snyk container test

# Docker Scout
$ go run main.go examples/job.yaml | xargs docker scout cves
```
