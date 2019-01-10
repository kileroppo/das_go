#!/usr/bin/env bash

cd $(dirname $0)

docker-compose down
if [[ -n $(docker images | grep das_go) ]]; then
    docker images | grep das_go | awk '{print $3 }' | xargs docker rmi
fi
docker pull registry.cn-hangzhou.aliyuncs.com/basicimage/golang:latest
docker-compose up -d --build

exit 0
