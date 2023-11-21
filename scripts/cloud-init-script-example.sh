#!/bin/bash

export NEW_PASS="123123"
export SSH_PUB_KEY="ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDDBk95x2NSixQpaRWTFIU3S1g9qUAz2BKBvT8aEOr37eJR4iOr1R8eJRH49M7t+TkBkXLDAX10da05y9tkkeP1wXFbtKsbh1KX5ILbMvY+FsmH4sUJ9pJTwfza0+BF3VIigG7TqDEPYxwj/j7+5yQI1Uru8vi/hakW/8uC1uZB9UlIveXLDtM9NrgRHbFcTmEY31ZtNaCUbs0C83N+xK2PBPky7N7S6hz95Gk29pwDB/OYuW4+GNj+3C7gmer2/QbUtVLbwPxqMxIR3JwvcJBJqaXcYWsOJRojwIyU0rdTZjUU+yngA6lPyOxz3Z/oCxt/NJ0Lb3Y4YnwQyP42CRYqYru/dJIcSWldZLZH8DEE6ZnY/UZ29kKW0mhrJpxSBp7+sDeUPw7aD3eHro65SgnDlbGW1nRtq/AmqaenWluuC0jfIF0YqTSi/2nNxDmvmZsKLtffvtWFIegnzn2+w0wo2LY6VhIx4eX9u6pyQ4qia6INVDkBM9X4AWcKh12ddrc= rostislav@rostislav"
export BOOTSTRAP_SSH_KEY="ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAEAQDgeaO3zkY5v1dww3fFONPzUnEgIqJ4kUK0Usu8iFdp+TWIulw9dDeQHa+PdWXP97l5Vv1mG9ipqShEIu7/2bp13KxSblWX4iV1MYZbtimhY3UDOsPn1G3E1Ipis6y+/tosDAm8LoWaGEMcLuE5UjP6gs6K57rCEjkVGjg7vjhIypAMC0J2N2+CxK9o/Y1+LZAec+FL5cmSJoajoo9y/VYJjz/a52jJ93wRafD2uu6ObAl5gkN/+gqY4IJFSMx20sPsIRXdbiBNDqiap56ibHHPKTeZhRbdXvZfjYkHtutXnSM2xn7BjnV8CguISxS3rXlzlzRVYwSUjnKUf5SKBbeyZbCokx271vCeUd5EXfHphvW6FIOna2AI5lpCSYkw5Kho3HaPi2NjXJ9i2dCr1zpcZpCiettDuaEjxR0Cl4Jd6PrAiAEZ0Ns0u2ysVhudshVzQrq6qdd7W9/MLjbDIHdTToNjFLZA6cbE0MQf18LXwJAl+G/QrXgcVaiopUmld+68JL89Xym55LzkMhI0NiDtWufawd/NiZ6jm13Z3+atvnOimdyuqBYeFWgbtcxjs0yN9v7h7PfPh6TbiQYFx37DCfQiIucqx1GWmMijmd7HMY6Nv3UvnoTUTSn4yz1NxhEpC61N+iAInZDpeJEtULSzlEMWlbzL4t5cF+Rm1dFdq3WpZt1mi8F4DgrsgZEuLGAw22RNW3++EWYFUNnJXaYyctPrMpWQktr4DB5nQGIHF92WR8uncxTNUXfWuT29O9e+bFYh1etmq8rsCoLcxN0zFHWvcECK55aE+47lfNAR+HEjuwYW10mGU/pFmO0F9FFmcQRSw4D4tnVUgl3XjKe3bBiTa4lUrzrKkLZ6n9/buW2e7c3jbjmXdPh2R+2Msr/vfuWs9glxQf+CYEbBW6Ye4pekIyI77SaB/bVhaHtXutKxm+QWdNle8aeqiA8Ji1Ml+s75vIg+n5v6viCnl5aV33xHRFpGQJzj2ktsXl9P9d5kgal9eXJYTywC2SnVbZVLb6FGN4kPZTVwX1f+u7v7JCm4YWlbQZtwwiXKjs99AVtQnBWqQvUH5sFUkVXlHA1Y9W6wlup0r+F6URL+7Yw+d0dHByfevrJg3pvmpLb3sEpjIAZodW3dIUReE7Ku3s/q/O9foFnfRBnCcZ2QsnxI5pqNrbrundD1ApOnNXEvICvPXHBBQ44cW0hzAO+WxY5VxyG8y/kXnb48G9efkIQFkNaITJrU9SiOk6bFP4QANdS/pmaSLjJIsHixa+7vmYjRy1SVoQ/39vDUnyCbqKtO56QMH32hQLRO3Vk7NVG6o4dYjFkiaMSaqVlHKMkJQHVzlK2PW9/fjVXfkAHmmhoD debian"
export SSH_POD="ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDNJoDynzwPT9VrgCdGG33xNhkboXNmDObjQTI88GQy+r/y9rwShvA+B+9+J9owuBdKuEnqk/CPFgTwbxHPS4SFoSnlS4dfL+stwVzQ0aTrI+Wz71xm91dVjS/N6xPIGCzSEZ3dzU+GNDObn5eZhlRx60S+zGI3Vj6siKszatCWhKR9I9ZyyJxArLTvhyY9uLCGDvBFoyLcCdwI8OjaZwBHn1HxkOc9Zi2roXZ01con5ABFFPADbit6AXjzfKfRjfh2gqfyTq9kXqGgqhF8ccTWA2YdPrxJfZ617FqJ1IvPzjzmSOp6ToIBP3UrGg0T+wC0TqFT+WDEj+udxW8y4y0hnA9HjzZEoEHoWcfz1LajqQY6FEb10+IHRj3BErn5SBUgvsrhrZlDoiKBWe4ZSLxdq0ElGWMs7uLz70YIkMuD8iAg2nIDRKfRLdnpmGfvneu55zP1edccdCmlT2emx+llO6fs5rzWS3vYOQkLM0qZBemuXvpPXUarp7CPCUenXU0= root@ssh"

CONFIGFILE="/var/lib/instaclustr/etc/cassandra/cassandra.yaml"

echo "debian:$NEW_PASS" | chpasswd
echo "core:$NEW_PASS" | sudo chpasswd core
echo "root:$NEW_PASS" | sudo chpasswd root

while true
do
  if [ -f "$CONFIGFILE" ]; then
    IPADDR=$(ip addr show enp1s0 | grep "inet\b" | awk '{print $2}' | cut -d/ -f1)
    sed -i "s/.*broadcast_rpc_address.*/broadcast_rpc_address: ${IPADDR}/g" $CONFIGFILE
    sed -i "s/.*broadcast_address.*/broadcast_address: ${IPADDR}/g" $CONFIGFILE
    sed -i "s/.*listen_address.*/listen_address: ${IPADDR}/g" $CONFIGFILE
    sed -i "s/.*{seeds.*/\ \ - {seeds: 'kubevirt-rostyslp-on-prem-cassandra.default.svc.cluster.local'}/g" $CONFIGFILE
    break
  else
    if [ -f "/ignition_done" ]; then
      sleep 5
    else
      sed -i 's/scripts-user$/\[scripts-user, always\]/' /etc/cloud/cloud.cfg
      sudo echo "$SSH_PUB_KEY" > /home/debian/.ssh/authorized_keys
      sudo echo "$BOOTSTRAP_SSH_KEY" >> /home/debian/.ssh/authorized_keys
      sudo echo "$SSH_POD" >> /home/debian/.ssh/authorized_keys
      sudo chown -R debian: /home/debian/.ssh
      sudo cp /usr/share/doc/apt/examples/sources.list /etc/apt/sources.list
      data_device=$(lsblk -dfn -o NAME,SERIAL | awk '$2 == "DATADISK" {print $1}')
      sudo mkfs -t ext4 /dev/"${data_device}"
      ignition_device=$(lsblk -dfn -o NAME,SERIAL | awk '$2 == "IGNITION" {print $1}')
      sudo mkdir /ignition
      sudo mount /dev/"${ignition_device}" /ignition/
      sudo cp /ignition/script /ignition.sh
      sudo chmod +x /ignition.sh
      /ignition.sh
      touch /ignition_done
    fi
  fi
done
