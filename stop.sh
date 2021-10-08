#!/bin/bash

sudo systemctl stop mjpg_streamer
sudo systemctl disable mjpg_streamer
sudo systemctl stop telesight
sudo systemctl disable telesight