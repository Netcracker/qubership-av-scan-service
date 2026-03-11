# Developer Guide

## How to update third-parties

To update third-parties, consider following:
1. Update base images in [`Dockerfile`](/build/Dockerfile). New versions could be found on docker hub:
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
1. Manually update [`appVersion` in `Chart.yaml`](/charts/av-scan-service/Chart.yaml) if needed
2. Decide on a new version, e.g. `0.10.1`
3. Run 'Helm Charts Release' workflow with specified version

**Note:** use your own applicable version instead of `0.10.1`

## GitHub PR labels

The following GitHub PR labels are used to group changes in release notes:

- breaking-change — use for changes that break backward compatibility, modify public API/contracts, or require user action during upgrade.
- feature — use for new product functionality or visible user-facing capabilities.
- enhancement — use for improvements to existing functionality, UX, or behavior.
- bug / fix / bugfix — use for defect resolution and regression fixes.
- refactor — use for internal code improvements that do not change user-facing behavior.
- documentation — use for docs-only updates, guides, READMEs, comments, or examples.

These labels affect only how changes are grouped in release notes.