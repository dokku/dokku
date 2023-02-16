#!/bin/bash
set -eo pipefail
set -o errexit

echo '--> Prevent login until after setup is complete'
# Be a bit of a dork about login in
cat >>/etc/ssh/sshd_config <<EOM
Match User root
        ForceCommand echo "Please wait while we get your droplet ready..."
EOM
