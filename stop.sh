#!/bin/bash

if service telesight status > /dev/null; then
    sudo systemctl stop telesight
    sudo systemctl disable telesight
fi
if service mjpg_streamer status > /dev/null; then
    sudo systemctl stop mjpg_streamer
    sudo systemctl disable mjpg_streamer
fi