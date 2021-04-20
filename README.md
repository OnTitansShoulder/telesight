# telesight (WIP)

A golang server that serves a UI frontend for the streams backed by the [mjpg-streamer](https://github.com/jacksonliam/mjpg-streamer) service. It is intended for local use, while accessing it remotely from a VPN service installed.

It supports viewing the streams live and view past and saved videos captured from the camera stream (using `ffmpeg`).

## Install

Ensure `git` is installed, then clone this repo and run `make install`.