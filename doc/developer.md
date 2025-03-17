# Developer Guide

## How to update third-parties

To update third-parties, consider following:
1. Update base images in [`Dockerfile`](/Dockerfile). New versions could be found on docker hub:
    * https://hub.docker.com/_/alpine
    * https://hub.docker.com/_/golang
2. Update go libraries. For this, in repository root, run following:
    ```
    go get -u
    go mod tidy
    ```
3. Update ClamAV version in [`/charts/av-scan-service/values.yaml`](/charts/av-scan-service/values.yaml).

## How to make release

To make release, do following:
1. Decide on a new version, e.g.  `0.6.2`
2. Manually update [`appVersion` in `Chart.yaml`](/charts/resource-reporter/Chart.yaml)
3. Create tag `0.6.2`

**Note:** use your own applicable version instead of `0.6.2`