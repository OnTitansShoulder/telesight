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
  libjpeg8-dev
  
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

# build the source for telepight
cd $CWD || (echo "cd failed" && exit 1)
go install
TELEPIGHT_BIN=$(which telepight)
[[ -x $TELEPIGHT_BIN ]] || (echo "Failed to install telepight" && exit 1)

# generate the services files
id -u telepight >> /dev/null || sudo useradd telepight
TELEPIGHT_RUN_ROOT=/run/telepight
sudo mkdir -p $TELEPIGHT_RUN_ROOT "$TELEPIGHT_RUN_ROOT/frames" $TELEPIGHT_RUN_ROOT/videos $TELEPIGHT_RUN_ROOT/templates
sudo cp $CWD/templates/*.gtpl "$TELEPIGHT_RUN_ROOT/templates/"
sudo cp $CWD/$STREAMER_ROOT/*.so $TELEPIGHT_RUN_ROOT
sudo cp -r "$CWD/$STREAMER_ROOT/www" "$TELEPIGHT_RUN_ROOT/"
sudo chown -R telepight:telepight $TELEPIGHT_RUN_ROOT
cat > telepight.service <<EOF
[Unit]
Description=camera streaming portal and video viewing/recording service
ConditionPathExists=$TELEPIGHT_RUN_ROOT

[Service]
User=root
Restart=always
RestartSec=5
WatchdogSec=21600
Nice=10

ExecStartPre=/bin/chown -R telepight:telepight $TELEPIGHT_RUN_ROOT
ExecStartPre=/bin/chmod -R 0755 $TELEPIGHT_RUN_ROOT

ExecStart=$TELEPIGHT_BIN -m $PRIMARY_HOSTNAME -b $TELEPIGHT_RUN_ROOT $STREAM_FLAG

[Install]
WantedBy=multi-user.target
EOF

cat > mjpg_streamer.service <<EOF
[Unit]
Description=webcam streaming service
ConditionPathExists=$TELEPIGHT_RUN_ROOT

[Service]
User=root
Restart=always
RestartSec=5
WatchdogSec=21600
Nice=10
WorkingDirectory=$TELEPIGHT_RUN_ROOT
ExecStartPre=/bin/chown -R telepight:telepight $TELEPIGHT_RUN_ROOT
ExecStartPre=/bin/chmod -R 0755 $TELEPIGHT_RUN_ROOT

ExecStart=$STREAMER_BIN -i 'input_uvc.so -r 640x360 -f 10' -o 'output_http.so'

[Install]
WantedBy=multi-user.target
EOF

sudo chmod 644 mjpg_streamer.service telepight.service
sudo cp mjpg_streamer.service telepight.service /lib/systemd/system/
sudo systemctl enable mjpg_streamer telepight
sudo systemctl start mjpg_streamer telepight
