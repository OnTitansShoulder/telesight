#!/bin/bash

CURRENT_IP=$(ifconfig | grep -Eo 'inet (addr:)?([0-9]{1,3}\.){3}[0-9]{1,3}' | grep -Eo '([0-9]{1,3}\.){3}[0-9]{1,3}' | grep -v '127.0.0.1')
echo "Current ip is $CURRENT_IP. We will update the eth0 interface with a new static ip address."
read -rp "What should be the static IP? > " NEW_STATIC_IP

function validate_ip {
    ip=$1
    for part in ${ip//\./ }; do
        if [[ $part -gt 255 ]]; then
            return 1
        fi
    done
    return 0
}

if [[ $NEW_STATIC_IP =~ ^([0-9]{1,3}\.){3}[0-9]{1,3}$ ]]; then
    if validate_ip "$NEW_STATIC_IP"; then
        echo "Setting $NEW_STATIC_IP as new static ip ..."
        if [[ -f /etc/dhcpcd.conf.telesight.bak ]]; then
            sudo cp /etc/dhcpcd.conf{.telesight.bak,}
        elif [[ -f /etc/dhcpcd.conf ]]; then
            sudo cp /etc/dhcpcd.conf{,.telesight.bak}
        fi

        sudo tee -a /etc/dhcpcd.conf > /dev/null <<EOF

# static profile
interface eth0
static ip_address=$NEW_STATIC_IP
static routers=192.168.1.1
static domain_name_servers=192.168.1.1
EOF
    fi
else
    echo "invalid ip address entered! ($NEW_STATIC_IP)"
    exit 1
fi

echo "Please reboot to let new static ip come in effect."