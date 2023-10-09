#!/bin/bash

path=$(readlink -f ../.env)

. $path
export USERNAME=$(echo -n $USERNAME | base64)
export APIKEY=$(echo -n $APIKEY | base64)
export HOSTNAME=$(echo -n $HOSTNAME | base64)
export ICADMIN_USERNAME=$(echo -n $ICADMIN_USERNAME | base64)
export ICADMIN_APIKEY=$(echo -n $ICADMIN_APIKEY | base64)

( echo "cat <<EOF >../config/manager/creds_secret.yaml";
  cat secret.yaml;
  echo "EOF";
) >tmp.yaml
. tmp.yaml
cat ../config/manager/creds_secret.yaml
rm tmp.yaml