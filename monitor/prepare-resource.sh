#!/usr/bin/env bash

resource_group="zuya$(date +'%Y%m%d')rg"
keyvault_name="zuya$(date +'%Y%m%d')akv"
storage_acount="zuya$(date +'%Y%m%d')account"
acr_name="zuya$(date +'%Y%m%d')acr"
cluster_name="zuya$(date +'%Y%m%d')aks"

exists=$(az group exists --name "$resource_group")
if [[ "$exists" == false ]]; then
    echo "create rg: $resource_group"
    az group create -n "$resource_group" -l westus2
else
    echo "skip create rg: $resource_group"
fi

az keyvault show --name "$keyvault_name" > /dev/null 2>&1
if [[ "$?" -ne 0 ]]; then
    echo "create akv: $keyvault_name"
    az keyvault create -n "$keyvault_name" -g "$resource_group"
else
    echo "skip create akv: $keyvault_name"
fi

az storage account show -n "$storage_acount" -g "$resource_group" > /dev/null 2>&1
if [[ "$?" -ne 0 ]]; then
    echo "create storage acount: $storage_acount"
    az storage account create -n "$storage_acount" -g "$resource_group"
else
    echo "skip create storage acount: $storage_acount"
fi

az acr show -n "$acr_name" -g "$resource_group" > /dev/null 2>&1
if [[ "$?" -ne 0 ]]; then
    echo "create acr acount: $acr_name"
    az acr create -n "$acr_name" -g "$resource_group" --sku Standard
else
    echo "skip create acr acount: $acr_name"
fi

az aks show -n "$cluster_name" -g "$resource_group" > /dev/null 2>&1
if [[ "$?" -ne 0 ]]; then
    echo "create aks cluster: $cluster_name"
    az aks create -n "$cluster_name" -g "$resource_group"
else
    echo "skip create aks cluster: $cluster_name"
fi

arc_id=$(az acr show -n "$acr_name" -g "$resource_group" --query "id" --output tsv)
az aks update -n "$cluster_name" -g "$resource_group" --attach-acr "$arc_id"
