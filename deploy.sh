#!/usr/bin/env bash

helm upgrade monitor --install \
  --wait \
  --timeout=600 \
  --namespace=monitor \
  "./charts/monitor" \
  --set image.registry="zuya20200930acr.azurecr.io" \
  --set-string image.tag="v1" \
  --set-string env.personalAccessToken="${PERSONAL_ACCESS_TOKEN}" \
  --set-string env.storageAccountAccessKey="${AZURE_STORAGE_ACCESS_KEY}" \
  --set-string azureStorageAccount="zuya20200930account" \
  --set-string azureCICDContainer="test"
