#!/usr/bin/env bash

hook_dir=/etc/letsencrypt/renewal-hooks

project_root="$(git rev-parse --show-toplevel)"
shebang="#!/bin/sh"


sudo cat << EOF > "$hook_dir/pre/tty-proxy.sh"
$shebang

docker-compose -f "$project_root/docker-compose.yml" down 
EOF

sudo cat << EOF > "$hook_dir/post/tty-proxy.sh"
$shebang

docker-compose -f "$project_root/docker-compose.yml" up -d 
EOF

sudo chmod 755 "$hook_dir/pre/tty-proxy.sh"
sudo chmod 755 "$hook_dir/post/tty-proxy.sh"

