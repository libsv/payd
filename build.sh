#!/bin/bash

env GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.commit=00a -X main.version=MANUAL"
