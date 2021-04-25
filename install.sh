#!/bin/bash

function installPackages() {
  # install prerequisite packages
  echo "Be ready to be prompted for sudo access to install packages..."
  sudo apt-get update
  sudo apt-get upgrade -y
  sudo apt-get install -y make gcc g++ golang haproxy
  if [[ $IS_STREAMING == "y" || $IS_STREAMING == "Y" ]]; then
    sudo apt-get install -y ffmpeg libjpeg8-dev
  fi
}

function build_streamer() {
  # build the source for streaming
  if [[ ! -d $STREAMER_ROOT ]]; then
    git clone https://github.com/jacksonliam/mjpg-streamer.git streamer-tmp
    mv streamer-tmp/mjpg-streamer-experimental $STREAMER_ROOT
    rm -rf streamer-tmp
  fi

  cd $STREAMER_ROOT || (echo "cd failed" && exit 1)
  make
  sudo make install
  cd $CWD || (echo "cd failed" && exit 1)
  [[ -x $(which mjpg_streamer) ]] || (echo "Failed to install mjpg_streamer" && exit 1)
}

function build_telesight() {
  # build the source for telesight
  cd $CWD || (echo "cd failed" && exit 1)
  go build
  sudo mv telesight /usr/local/bin/
  [[ -x $(which telesight) ]] || (echo "Failed to install telesight" && exit 1)
}

function generate_telesight_service() {
  # generate the services files
  id -u telesight >> /dev/null || sudo useradd telesight
  TELESIGHT_RUN_ROOT=/var/telesight
  sudo mkdir -p $TELESIGHT_RUN_ROOT "$TELESIGHT_RUN_ROOT/frames" $TELESIGHT_RUN_ROOT/videos $TELESIGHT_RUN_ROOT/templates
  sudo cp $CWD/templates/*.gtpl "$TELESIGHT_RUN_ROOT/templates/"
  sudo chown -R telesight:telesight $TELESIGHT_RUN_ROOT
  sudo chmod -R 0755 $TELESIGHT_RUN_ROOT
  cat > telesight.service <<EOF
[Unit]
Description=camera streaming portal and video viewing/recording service
ConditionPathExists=$TELESIGHT_RUN_ROOT

[Service]
User=root
Restart=always
RestartSec=5
WatchdogSec=21600
Nice=10
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=telesight

ExecStart=$TELESIGHT_BIN -m $PRIMARY_HOSTNAME -b $TELESIGHT_RUN_ROOT $STREAM_FLAG -r

[Install]
WantedBy=multi-user.target
EOF

  cat > telesight.conf <<EOF
:programname, startswith, "telesight" {
  /var/log/telesight.log
  stop
}
EOF

  sudo chmod 644 telesight.service telesight.conf
  sudo cp telesight.conf /etc/rsyslog.d/
  sudo cp telesight.service /lib/systemd/system/
  sudo systemctl enable telesight
  sudo systemctl restart telesight rsyslog
}

function generate_streamer_service() {
  getSupportedVideoFormat

  TELESIGHT_RUN_ROOT=/var/telesight
  sudo mkdir -p $TELESIGHT_RUN_ROOT
  sudo cp $STREAMER_ROOT/*.so $TELESIGHT_RUN_ROOT
  sudo cp -r "$STREAMER_ROOT/www" "$TELESIGHT_RUN_ROOT/"
  sudo chown -R telesight:telesight $TELESIGHT_RUN_ROOT
  sudo chmod -R 0755 $TELESIGHT_RUN_ROOT
  cat > mjpg_streamer.service <<EOF
[Unit]
Description=webcam streaming service
ConditionPathExists=$TELESIGHT_RUN_ROOT

[Service]
User=root
Restart=always
RestartSec=5
WatchdogSec=21600
Nice=10
WorkingDirectory=$TELESIGHT_RUN_ROOT
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=mjpg_streamer

ExecStart=$STREAMER_BIN -i 'input_uvc.so -r $RESOLUTION -f 10 $FORMAT_FLAG' -o 'output_http.so'

[Install]
WantedBy=multi-user.target
EOF

  cat > mjpg_streamer.conf <<EOF
:programname, startswith, "mjpg_streamer" {
  /var/log/mjpg_streamer.log
  stop
}
EOF

  sudo chmod 644 mjpg_streamer.service mjpg_streamer.conf
  sudo cp mjpg_streamer.conf /etc/rsyslog.d/
  sudo cp mjpg_streamer.service /lib/systemd/system/
  sudo systemctl enable mjpg_streamer
  sudo systemctl restart mjpg_streamer rsyslog
}

function getSupportedVideoFormat() {
  VIDEO_FORMATS=$(ffmpeg -f v4l2 -list_formats all -i /dev/video0 2>&1 | grep 'v4l2')
  if [[ -n $(echo $VIDEO_FORMATS | grep 'mjpeg') ]]; then
    SUPPORT_MJPG=true
    MJPG_RESOLUTIONS=$(echo $VIDEO_FORMATS | grep 'mjpeg' | rev | cut -d':' -f 1 | rev | awk '{$1=$1;print}')
    getResolution "$MJPG_RESOLUTIONS"
    MJPG_RESOLUTION=$RESOLUTION
  fi
  if [[ -n $(echo $VIDEO_FORMATS | grep 'yuyv') ]]; then
    SUPPORT_YUV=true
    YUV_RESOLUTIONS=$(echo $VIDEO_FORMATS | grep 'yuyv' | rev | cut -d':' -f 1 | rev | awk '{$1=$1;print}')
    getResolution "$YUV_RESOLUTIONS"
    YUV_RESOLUTION=$RESOLUTION
  fi
  if [[ $SUPPORT_MJPG == 'true' ]]; then
    FORMAT_FLAG=
    RESOLUTION=$MJPG_RESOLUTION
  elif [[ $SUPPORT_YUV == 'true' ]]; then
    FORMAT_FLAG="-y yuv"
    RESOLUTION=$YUV_RESOLUTION
  else
    echo "No supported video format found from the cam /dev/video0" && exit 1
  fi
}

function getResolution() {
  RESOLUTIONS=$1
  if [[ -n $(echo $RESOLUTIONS | grep '640x360') ]]; then
    RESOLUTION='640x360'
  elif [[ -n $(echo $RESOLUTIONS | grep '640x480') ]]; then
    RESOLUTION='640x480'
  else
    RESOLUTION=$(echo $RESOLUTIONS | cut -d' ' -f 1)
  fi
}

function update_haproxy_config() {
  # setup haproxy to route the traffic
  if [[ -e /etc/haproxy/haproxy.cfg.telesight.bak ]]; then
    sudo cp /etc/haproxy/haproxy.cfg{.telesight.bak,}
  else
    sudo cp /etc/haproxy/haproxy.cfg{,.telesight.bak}
  fi

  sudo tee -a /etc/haproxy/haproxy.cfg <<EOF
frontend public
        bind :::80 v4v6
        option forwardfor except 127.0.0.1
        use_backend telesight if { path_beg /telesight/ }
EOF

  if [[ $IS_STREAMING == "y" || $IS_STREAMING == "Y" ]]; then
    sudo tee -a /etc/haproxy/haproxy.cfg <<EOF
        use_backend webcam if { path_beg /webcam/ }
EOF
  fi
  sudo tee -a /etc/haproxy/haproxy.cfg <<EOF

backend telesight
        reqrep ^([^\ :]*)\ /telesight/(.*)     \1\ /\2
        server telesight  127.0.0.1:8089
EOF
  if [[ $IS_STREAMING == "y" || $IS_STREAMING == "Y" ]]; then
  sudo tee -a /etc/haproxy/haproxy.cfg <<EOF
backend webcam
        reqrep ^([^\ :]*)\ /webcam/(.*)     \1\ /\2
        server webcam1  127.0.0.1:8080
EOF
  fi
  sudo systemctl restart haproxy
}

# collect necessary info upfront
read -p "Is this the first time running this script on this host (y/N) > " IS_FRESH_INSTALL

echo "Enter the primary instance's hostname (without .local)"
read -p "or press enter to use current machine as primary > " PRIMARY_HOSTNAME
[[ -n $PRIMARY_HOSTNAME ]] || PRIMARY_HOSTNAME=$(hostname)

read -p "Does this instance have a webcam and will be serving a stream (y/N) > " IS_STREAMING
[[ $IS_STREAMING == "y" || $IS_STREAMING == "Y" ]] && STREAM_FLAG="-s"


CWD=$(pwd)
STREAMER_ROOT="$CWD/mjpg-streamer"
if [[ $IS_FRESH_INSTALL == "y" || $IS_FRESH_INSTALL == "Y" ]]; then
  installPackages
  if [[ $IS_STREAMING == "y" || $IS_STREAMING == "Y" ]]; then
    build_streamer
  fi
  build_telesight
fi

TELESIGHT_BIN=$(which telesight)
[[ -x $TELESIGHT_BIN ]] || (echo "Failed to install telesight" && exit 1)

if [[ $IS_STREAMING == "y" || $IS_STREAMING == "Y" ]]; then
  STREAMER_BIN=$(which mjpg_streamer)
  [[ -x $STREAMER_BIN ]] || (echo "Failed to install mjpg_streamer" && exit 1)
fi

if [[ $IS_STREAMING == "y" || $IS_STREAMING == "Y" ]]; then
  generate_streamer_service
fi
generate_telesight_service
update_haproxy_config
