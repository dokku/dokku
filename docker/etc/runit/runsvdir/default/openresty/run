#!/bin/bash

exec >>/var/log/services/openresty
exec 2>&1

echo "Running openresty"
exec /usr/bin/openresty -c /usr/local/openresty/nginx/conf/nginx.conf -g "daemon off;"
