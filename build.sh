#!/bin/bash

CGO_ENABLED=0 go build -o build/k8s-secret-creator -v main.go
