#!/bin/bash
set -xe

GOOS=linux GOARCH=arm go build -v
