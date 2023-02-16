#!/bin/sh
#
# Configured as part of the DigitalOcean 1-Click Image build process

myip=$(hostname -I | awk '{print$1}')
cat <<EOF
********************************************************************************

Welcome to DigitalOcean's 1-Click Dokku Droplet.
To keep this Droplet secure, the UFW firewall is enabled. 
All ports are BLOCKED except 22 (SSH), 80 (HTTP), 443 (HTTPS).

In a web browser, you can view:
 * The Dokku One-Click Quickstart guide: https://do.co/dokku1804#start

For help and more information, visit https://do.co/dokku1804

********************************************************************************
To delete this message of the day: rm -rf $(readlink -f ${0})
To remove the default web landing page: rm -rf /etc/nginx/sites-available/digitalocean
EOF
