#!/usr/bin/env bash

docker build . -t monitor:v1
docker tag monitor:v1 zuya20201002acr.azurecr.io/monitor:v1

az acr login -n zuya20201002acr
docker push zuya20201002acr.azurecr.io/monitor:v1
