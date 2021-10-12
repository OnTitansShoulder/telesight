#!/bin/bash

echo "Starting mjpg_streamer ..."

sudo systemctl enable mjpg_streamer
sudo systemctl restart mjpg_streamer

sleep 5

echo "Starting telesight ..."

sudo systemctl enable telesight
sudo systemctl restart telesight rsyslog