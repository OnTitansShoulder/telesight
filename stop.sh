#!/bin/bash

if systemctl list-units --full -all | grep telesight | awk '{printf "%s %s %s %s\n", $1,$2,$3,$4}'; then
    sudo systemctl stop telesight
    sudo systemctl disable telesight
fi
if systemctl list-units --full -all | grep mjpg_streamer | awk '{printf "%s %s %s %s\n", $1,$2,$3,$4}'; then
    sudo systemctl stop mjpg_streamer
    sudo systemctl disable mjpg_streamer
fi