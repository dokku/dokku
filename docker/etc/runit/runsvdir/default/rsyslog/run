#!/bin/bash

exec >>/var/log/services/rsyslog
exec 2>&1

echo "Running rsyslog"
exec /usr/sbin/rsyslogd -n
