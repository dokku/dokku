#!/bin/bash

exec >>/var/log/services/dokku-restore
exec 2>&1

if [[ "$1" -ne 0 ]]; then
  echo "Error restoring dokku apps! Retrying in 30 seconds..."
  exec sleep 30
  exit 0
fi

echo "Done restoring dokku apps."
exec sleep infinity

exit 0
