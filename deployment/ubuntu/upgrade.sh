#!/bin/bash

function sloth {
    echo "Upgrade sloth ..."

    cd /etc/sloth
    mv config.prod.yml config.back.yml

    LATEST_VERSION=$(curl --silent "https://api.github.com/repos/Clivern/Sloth/releases/latest" | jq '.tag_name' | sed -E 's/.*"([^"]+)".*/\1/' | tr -d v)

    curl -sL https://github.com/Clivern/Sloth/releases/download/v{$LATEST_VERSION}/sloth_{$LATEST_VERSION}_Linux_x86_64.tar.gz | tar xz

    rm config.prod.yml
    mv config.back.yml config.prod.yml

    systemctl restart sloth

    echo "Sloth upgrade done!"
}

sloth
