#!/bin/bash

function deps {
    echo "Installing dependencies ..."

    apt-get update
    apt-get upgrade -y
    apt-get install stress jq -y
    apt-get install net-tools -y

    echo "Installing dependencies done!"
}

function chipmunk {
    echo "Installing chipmunk ..."

    mkdir -p /etc/chipmunk
    cd /etc/chipmunk
    LATEST_VERSION=$(curl --silent "https://api.github.com/repos/clivern/chipmunk/releases/latest" | jq '.tag_name' | sed -E 's/.*"([^"]+)".*/\1/' | tr -d v)
    curl -sL https://github.com/clivern/chipmunk/releases/download/v{$LATEST_VERSION}/chipmunk_{$LATEST_VERSION}_Linux_x86_64.tar.gz | tar xz

    echo "[Unit]
Description=Chipmunk
Documentation=https://github.com/clivern/Chipmunk

[Service]
ExecStart=/etc/chipmunk/chipmunk server -c /etc/chipmunk/config.prod.yml
Restart=on-failure
RestartSec=2

[Install]
WantedBy=multi-user.target" > /etc/systemd/system/chipmunk.service

    systemctl daemon-reload
    systemctl enable chipmunk.service
    systemctl start chipmunk.service

    echo "Chipmunk installation done!"
}

deps
chipmunk
