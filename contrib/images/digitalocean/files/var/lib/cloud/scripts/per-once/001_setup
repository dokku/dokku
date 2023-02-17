#!/bin/bash -x
set +e
exec > >(tee /var/log/one_click_setup.log) 2>&1

ufw --force enable
printf "dokku dokku/hostname string %s" hostname | debconf-set-selections

export DEBIAN_FRONTEND=noninteractive
export LANG=C
export LC_ALL=C

# Install Dokku....
for count in {1..30}; do
  dpkg -i /var/lib/digitalocean/debs/*deb || /bin/true
  apt-get -f install || /bin/true

  version=$(dpkg-query --showformat='${Version}' -W dokku)
  if [[ -n "$version" ]]; then
    ran=0
    exit 0
  else
    count=$((count + 1))
  fi
  sleep 5
done
