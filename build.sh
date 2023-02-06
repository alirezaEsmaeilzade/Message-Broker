#!/bin/bash
echo "---Building"
go build -o build/server cmd/server/server.go
echo "---Building docker image"
docker build . -t docker.bale.ai/bale/broker-alireza/broker:1.5 #must add "$1" for tag
echo "---Pushing docker image"
docker push docker.bale.ai/bale/broker-alireza/broker:1.5 #must add "$1" for tag
echo "Push done"
