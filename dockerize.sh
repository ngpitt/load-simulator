#!/bin/bash
set -xe

docker build -t ngpitt/load-simulator:v1 .
docker push ngpitt/load-simulator
