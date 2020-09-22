#!/usr/bin/env bash

docker build . -t monitor:v1
docker tag monitor:v1 zuya20200921acr.azurecr.io/monitor:v1

az acr login -n zuya20200921acr
docker push zuya20200921acr.azurecr.io/monitor:v1
