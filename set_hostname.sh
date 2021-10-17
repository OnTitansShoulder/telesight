#!/bin/bash

OLD_HOSTNAME=$(hostname)
read -rp "What should be the new hostname? > " NEW_HOSTNAME

if [[ $NEW_HOSTNAME =~ ^[0-9a-zA-Z]+(\.[0-9a-zA-Z]+)*$ ]]; then
    echo "Setting $NEW_HOSTNAME as new hostname ..."
    sudo hostnamectl --transient set-hostname "$NEW_HOSTNAME"
    sudo hostnamectl --static set-hostname "$NEW_HOSTNAME"
    sudo hostnamectl --pretty set-hostname "$NEW_HOSTNAME"
    sudo sed -i s/"$OLD_HOSTNAME"/"$NEW_HOSTNAME"/g /etc/hosts
else
    echo "invalid hostname entered! ($NEW_HOSTNAME)"
    exit 1
fi
