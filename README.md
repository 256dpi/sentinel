# sentinel

[![Release](https://img.shields.io/github/release/256dpi/sentinel.svg)](https://github.com/256dpi/sentinel/releases)
[![Docker Pulls](https://img.shields.io/docker/pulls/256dpi/sentinel.svg)](https://hub.docker.com/r/256dpi/sentinel)
[![Go Report Card](https://goreportcard.com/badge/github.com/256dpi/sentinel)](http://goreportcard.com/report/256dpi/sentinel)

**Kubernetes Event Reporter for Sentry**

Run `sentinel` inside your kubernetes cluster and configure it with the `SENTRY_DSN` of the project that should receive all kubernetes events:

```bash
kubectl run sentinel --image 256dpi/sentinel --env="SENTRY_DSN=..."
```
