#!/bin/bash
set -eo pipefail
set -o errexit

echo '--> Update grub boot loader'
sed -e 's|GRUB_CMDLINE_LINUX="|GRUB_CMDLINE_LINUX="cgroup_enable=memory swapaccount=1|g' \
  -i /etc/default/grub

update-grub
