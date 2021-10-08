#!/bin/bash

WD=$(pwd)
sudo apt-get update && sudo apt-get upgrade

# ensure required packages are installed
sudo apt-get install -y \
    cmake gcc g++ \
    git \
    golang \
    haproxy \
    libjpeg8-dev

# pull down mjpg-streamer (forked from jacksonliam) and build from source
git pull https://github.com/OnTitansShoulder/mjpg-streamer && \
    cd mjpg-streamer && \
    mv mjpg-streamer-experimental/* ./
    make && sudo make install

cd $WD

# TODO: install IPsec VPN from https://github.com/hwdsl2/setup-ipsec-vpn
