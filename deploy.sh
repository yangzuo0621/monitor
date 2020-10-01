#!/usr/bin/env bash

helm upgrade monitor --install \
  --wait \
  --timeout=600 \
  --namespace=monitor \
  "./charts/monitor" \
  --set image.registry="zuya20201002acr.azurecr.io" \
  --set-string image.tag="v1" \
  --set-string env.personalAccessToken="${PERSONAL_ACCESS_TOKEN}" \
  --set-string env.storageAccountAccessKey="${AZURE_STORAGE_ACCESS_KEY}" \
  --set-string azureStorageAccount="zuya20201002account" \
  --set-string repositoryURL="https://dev.azure.com/msazure/CloudNativeCompute/_git/aks-rp" \
  --set-string azureCICDContainer="test"
