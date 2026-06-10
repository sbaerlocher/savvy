#!/usr/bin/env bash
# @command observability
# @description Start observability stack (Grafana, Prometheus, Loki, Tempo)
docker compose --profile observability up -d "$@"
