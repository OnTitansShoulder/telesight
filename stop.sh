#!/bin/bash

if service telesight status > /dev/null; then
    sudo systemctl stop telesight
    sudo systemctl disable telesight
fi

if service mjpg_streamer status > /dev/null; then
    sudo systemctl stop mjpg_streamer
    sudo systemctl disable mjpg_streamer
fi

if service curl_echo_server status > /dev/null; then
    sudo systemctl stop curl_echo_server
    sudo systemctl disable curl_echo_server
fi