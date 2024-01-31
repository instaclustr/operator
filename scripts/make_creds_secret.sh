#!/bin/bash

path=$(readlink -f ~/.env)

. $path
export USERNAME=$(echo -n $USERNAME | base64)
export APIKEY=$(echo -n $APIKEY | base64)
export HOSTNAME=$(echo -n $HOSTNAME | base64)

( echo "cat <<EOF >~/creds_secret.yaml";
  cat secret.yaml;
  echo "EOF";
) >tmp.yaml
. tmp.yaml
cat ~/creds_secret.yaml
rm tmp.yaml