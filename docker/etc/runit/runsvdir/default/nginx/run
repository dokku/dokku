#!/bin/bash

exec >>/var/log/services/nginx
exec 2>&1

echo "Running nginx"
exec /usr/sbin/nginx -c /etc/nginx/nginx.conf -g "daemon off;"
