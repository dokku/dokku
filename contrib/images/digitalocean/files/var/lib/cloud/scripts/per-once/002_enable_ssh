#!/bin/bash
exec > >(tee /var/log/one_click_setup.log) 2>&1

# Remove the force command
sed -e '/Match user root/d' \
  -e '/.*ForceCommand.*droplet.*/d' \
  -i /etc/ssh/sshd_config

systemctl restart ssh
