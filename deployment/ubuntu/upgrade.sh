#!/bin/bash

function chipmunk {
    echo "Upgrade chipmunk ..."

    cd /etc/chipmunk
    mv config.prod.yml config.back.yml

    LATEST_VERSION=$(curl --silent "https://api.github.com/repos/Clivern/Chipmunk/releases/latest" | jq '.tag_name' | sed -E 's/.*"([^"]+)".*/\1/' | tr -d v)

    curl -sL https://github.com/Clivern/Chipmunk/releases/download/v{$LATEST_VERSION}/chipmunk_{$LATEST_VERSION}_Linux_x86_64.tar.gz | tar xz

    rm config.prod.yml
    mv config.back.yml config.prod.yml

    systemctl restart chipmunk

    echo "Chipmunk upgrade done!"
}

chipmunk
