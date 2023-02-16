#!/bin/bash
set -eo pipefail
set -o errexit

echo '--> Enable and configure UFW'
sed -e 's|DEFAULT_FORWARD_POLICY=.*|DEFAULT_FORWARD_POLICY="ACCEPT"|g' \
  -i /etc/default/ufw

ufw limit ssh
ufw allow 'Nginx Full'

ufw --force enable
