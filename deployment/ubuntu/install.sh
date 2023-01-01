#!/bin/bash

function deps {
    echo "Installing dependencies ..."

    apt-get update
    apt-get upgrade -y
    apt-get install stress jq -y
    apt-get install net-tools -y

    echo "Installing dependencies done!"
}

function sloth {
    echo "Installing sloth ..."

    mkdir -p /etc/sloth
    cd /etc/sloth
    LATEST_VERSION=$(curl --silent "https://api.github.com/repos/clivern/sloth/releases/latest" | jq '.tag_name' | sed -E 's/.*"([^"]+)".*/\1/' | tr -d v)
    curl -sL https://github.com/clivern/sloth/releases/download/v{$LATEST_VERSION}/sloth_{$LATEST_VERSION}_Linux_x86_64.tar.gz | tar xz

    echo "[Unit]
Description=Sloth
Documentation=https://github.com/clivern/Sloth

[Service]
ExecStart=/etc/sloth/sloth server -c /etc/sloth/config.prod.yml
Restart=on-failure
RestartSec=2

[Install]
WantedBy=multi-user.target" > /etc/systemd/system/sloth.service

    systemctl daemon-reload
    systemctl enable sloth.service
    systemctl start sloth.service

    echo "Sloth installation done!"
}

deps
sloth
