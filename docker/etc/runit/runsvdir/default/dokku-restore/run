#!/bin/bash

exec >>/var/log/services/dokku-restore
exec 2>&1

echo "Running dokku-restore"
cd /tmp || exit 1

exec dokku ps:restore
