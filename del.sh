#!/usr/bin/env bash

cd $(dirname $0)

docker-compose down
if [[ -n $(docker images | grep das_java) ]]; then
    docker images | grep das_go | awk '{print $3 }' | xargs docker rmi
fi

exit 0
