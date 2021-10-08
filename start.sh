#!/bin/bash

sudo systemctl enable mjpg_streamer
sudo systemctl restart mjpg_streamer
sudo systemctl enable telesight
sudo systemctl restart telesight rsyslog