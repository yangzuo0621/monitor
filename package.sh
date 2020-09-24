#!/usr/bin/env bash

docker build . -t monitor:v1
docker tag monitor:v1 zuya20200924acr.azurecr.io/monitor:v1

az acr login -n zuya20200924acr
docker push zuya20200924acr.azurecr.io/monitor:v1
