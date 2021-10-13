# set up wireless connection through command line

If running a Raspberry Pi server, you can do `raspi-config` to set the wireless connection.

Otherwise, you will need a program called `wpasupplicant` to help handle the wireless connection.

```sh
# for Ubuntu or Debian systems
sudo apt-get install wpasupplicant
```

Then update the config file at `/etc/wpa_supplicant/wpa_supplicant.conf`

```sh
sudo cat >> /etc/wpa_supplicant/wpa_supplicant.conf <<EOF
network={
    ssid="<ssid_name>"
    psk="<password>"
}
EOF
```

Start the wireless connection with

```sh
# you might need to replace the driver nl80211 with other drivers
# and replace the interface name wlan0 with the name of the wireless device interface name
sudo wpa_supplicant -Dnl80211 -iwlan0 -c/etc/wpa_supplicant/wpa_supplicant.conf
```
