# telesight

A golang server that serves a UI frontend for the streams backed by the [mjpg-streamer](https://github.com/jacksonliam/mjpg-streamer) service and `ffmpeg`.

It is intended for setting up a system of cameras with the ability to stream while also saving frames to filesystem as .mp4 videos. You will be able to access all camera streams and view past footages from the single web page.

Ideally you want to have several computing devices such as raspberry pis with some webcams supporting motion JPEG (mJEPG). A VPN need to be installed on one of the device and serve as the gateway to access your camera system.

Currently this program only supports Ubuntu and Debian OS.

## Install

Ensure `git` is installed, then clone this repo and run

```sh
make install
```

Follow the instructions from there. Be sure to setup the primary host first (which will serve the web page / UI), then provide its ip address for the rest of other camera hosts during the setup.