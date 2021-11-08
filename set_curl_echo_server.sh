#!/bin/bash

function generate_curlechoserver_service() {
  # generate the service files
  echo
  read -rp "What is the echo server hostname (without http:// or https://) > " ECHO_SERVER

  if [[ $(curl -w '%{http_code}' https://$ECHO_SERVER/health/ -o /dev/null -s) != "200" ]]; then
    echo "unable to reach the echo server at https://$ECHO_SERVER/health/, skip adding curl_echo_server service."
    echo
    return
  fi
  echo
  
  ECHO_URL=https://$ECHO_SERVER/echo/
  sudo cp "$PWD"/extras/curl_echo_server /usr/local/bin
  sudo chmod 755 /usr/local/bin/curl_echo_server
  cat > curl_echo_server.service <<EOF
[Unit]
Description=curl the echo server repeatedly to report current server's public ip address

[Service]
User=root
Restart=always
RestartSec=5
WatchdogSec=21600
Nice=10
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=curl_echo_server

ExecStart=/usr/local/bin/curl_echo_server $ECHO_URL

[Install]
WantedBy=multi-user.target
EOF

    sudo chmod 644 curl_echo_server.service
    sudo cp curl_echo_server.service /lib/systemd/system/
    sudo systemctl enable curl_echo_server
    sudo systemctl restart curl_echo_server
}

# check if we should add the curl_echo_server service
echo
echo "The echo server is meant to come up and serving as a reference"
echo "  to figure out the IP address of a remote server, in case the"
echo "  remote isn't configured to be attached to a domain name and"
echo "  its IP address is given by the ISP which can change over time."
echo 
echo "If you have setup an echo server using https://github.com/OnTitansShoulder/echo-server"
read -rp "  you can proceed [y/N] > " PROCEED

if [[ $PROCEED == "y" || $PROCEED == "Y" ]]; then
    generate_curlechoserver_service
fi
