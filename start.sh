#!/bin/bash

if service mjpg_streamer status > /dev/null; then
    echo "Starting mjpg_streamer ..."

    sudo systemctl enable mjpg_streamer
    sudo systemctl restart mjpg_streamer

    sleep 5
fi

if service telesight status > /dev/null; then
    echo "Starting telesight ..."

    sudo systemctl enable telesight
    sudo systemctl restart telesight rsyslog
fi

if service curl_echo_server status > /dev/null; then
    echo "Starting curl_echo_server ..."

    sudo systemctl enable curl_echo_server
    sudo systemctl restart curl_echo_server
fi