#!/bin/bash

export NEW_PASS="123123"
export SSH_PUB_KEY=""
export BOOTSTRAP_SSH_KEY=""

echo "debian:$NEW_PASS" | chpasswd
echo "root:$NEW_PASS" | sudo chpasswd root
sudo echo "$SSH_PUB_KEY" > /home/debian/.ssh/authorized_keys
sudo echo "$BOOTSTRAP_SSH_KEY" >> /home/debian/.ssh/authorized_keys
sudo chown -R debian: /home/debian/.ssh
sudo cp /usr/share/doc/apt/examples/sources.list /etc/apt/sources.list
data_device=$(lsblk -dfn -o NAME,SERIAL | awk '$2 == "DATADISK" {print $1}')
sudo mkfs -t ext4 /dev/${data_device}
ignition_device=$(lsblk -dfn -o NAME,SERIAL | awk '$2 == "IGNITION" {print $1}')
sudo mkdir /ignition
sudo mount /dev/${ignition_device} /ignition/
sudo cp /ignition/script /ignition.sh
sudo chmod +x /ignition.sh
/ignition.sh
END
