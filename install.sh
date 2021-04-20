#!/bin/bash

# collect necessary info upfront
echo "Enter the primary instance's hostname (without .local)"
read -p "or press enter to use current machine as primary > " PRIMARY_HOSTNAME
[[ -n $PRIMARY_HOSTNAME ]] || PRIMARY_HOSTNAME=$(hostname)

read -p "Does this instance have a webcam and will be serving a stream (y/N) > " IS_STREAMING
[[ $IS_STREAMING == "y" || $IS_STREAMING == "Y" ]] && STREAM_FLAG="-s"

# install prerequisite packages
echo "Be ready to be prompted for sudo access to install packages..."
sudo apt-get update
sudo apt-get upgrade -y
sudo apt-get install -y \
  cmake gcc g++ golang \
  ffmpeg \
  libjpeg8-dev \
  haproxy

# build the source for streaming
CWD=$(pwd)
STREAMER_ROOT=mjpg-streamer
git clone https://github.com/jacksonliam/mjpg-streamer.git streamer-tmp
mv streamer-tmp/mjpg-streamer-experimental $STREAMER_ROOT
rm -rf streamer-tmp
cd $STREAMER_ROOT || (echo "cd failed" && exit 1)
make
sudo make install
cd $CWD || (echo "cd failed" && exit 1)
STREAMER_BIN=$(which mjpg_streamer)
[[ -x $STREAMER_BIN ]] || (echo "Failed to install mjpg_streamer" && exit 1)

# build the source for telesight
cd $CWD || (echo "cd failed" && exit 1)
go build
sudo mv telesight /usr/local/bin/
TELESIGHT_BIN=$(which telesight)
[[ -x $TELESIGHT_BIN ]] || (echo "Failed to install telesight" && exit 1)

# generate the services files
id -u telesight >> /dev/null || sudo useradd telesight
TELESIGHT_RUN_ROOT=/run/telesight
sudo mkdir -p $TELESIGHT_RUN_ROOT "$TELESIGHT_RUN_ROOT/frames" $TELESIGHT_RUN_ROOT/videos $TELESIGHT_RUN_ROOT/templates
sudo cp $CWD/templates/*.gtpl "$TELESIGHT_RUN_ROOT/templates/"
sudo cp $CWD/$STREAMER_ROOT/*.so $TELESIGHT_RUN_ROOT
sudo cp -r "$CWD/$STREAMER_ROOT/www" "$TELESIGHT_RUN_ROOT/"
sudo chown -R telesight:telesight $TELESIGHT_RUN_ROOT
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

ExecStartPre=/bin/chown -R telesight:telesight $TELESIGHT_RUN_ROOT
ExecStartPre=/bin/chmod -R 0755 $TELESIGHT_RUN_ROOT

ExecStart=$TELESIGHT_BIN -m $PRIMARY_HOSTNAME -b $TELESIGHT_RUN_ROOT $STREAM_FLAG

[Install]
WantedBy=multi-user.target
EOF

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
ExecStartPre=/bin/chown -R telesight:telesight $TELESIGHT_RUN_ROOT
ExecStartPre=/bin/chmod -R 0755 $TELESIGHT_RUN_ROOT

ExecStart=$STREAMER_BIN -i 'input_uvc.so -r 640x360 -f 10' -o 'output_http.so'

[Install]
WantedBy=multi-user.target
EOF

sudo chmod 644 mjpg_streamer.service telesight.service
sudo cp mjpg_streamer.service telesight.service /lib/systemd/system/
sudo systemctl enable mjpg_streamer telesight
sudo systemctl start mjpg_streamer telesight

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
        use_backend webcam if { path_beg /webcam/ }

backend telesight
        reqrep ^([^\ :]*)\ /(.*)     \1\ /\2
        server telesight  127.0.0.1:8089

backend webcam
        reqrep ^([^\ :]*)\ /webcam/(.*)     \1\ /\2
        server webcam1  127.0.0.1:8080
EOF
sudo service haproxy restart