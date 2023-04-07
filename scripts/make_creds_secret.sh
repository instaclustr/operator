#!/bin/bash

path=$(readlink -f ../.env)

. $path
export USERNAME=$(echo -n $USERNAME | base64 -w0)
export APIKEY=$(echo -n $APIKEY | base64 -w0)
export HOSTNAME=$(echo -n $HOSTNAME | base64 -w0)
export CASSANDRA_DATA_VOLUME_IMAGE_URL=$(echo -n $CASSANDRA_DATA_VOLUME_IMAGE_URL | base64 -w0)

( echo "cat <<EOF >../config/manager/creds_secret.yaml";
  cat secret.yaml;
  echo "EOF";
) >tmp.yaml
. tmp.yaml
cat ../config/manager/creds_secret.yaml
rm tmp.yaml