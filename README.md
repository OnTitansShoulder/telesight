# telesight (WIP)

A golang server that serves a UI frontend for the streams backed by the [mjpg-streamer](https://github.com/jacksonliam/mjpg-streamer) service and `ffmpeg`.

It is intended for setting up a system of cameras with the ability to stream while also saving frames to filesystem as .mp4 videos.

Ideally you want to have several computing devices such as raspberry pis with some webcams supporting motion JPEG (mJEPG). A VPN need to be installed on one of the device and serve as the gateway to access your camera system.

Currently this codebase only supports Ubuntu OS.

## Install

Ensure `git` is installed, then clone this repo and run

```sh
make preinstall
make install
```