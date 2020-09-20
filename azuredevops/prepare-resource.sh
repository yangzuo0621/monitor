#!/usr/bin/env bash

resource_group="zuya$(date +'%Y%m%d')rg"
keyvault_name="zuya$(date +'%Y%m%d')akv"
storage_acount="zuya$(date +'%Y%m%d')account"

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
