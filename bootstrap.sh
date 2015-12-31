#!/bin/bash
#
# A script to bootstrap dokku.
# It expects to be run on Ubuntu 14.04 via 'sudo'
# If installing a tag higher than 0.3.13, it may install dokku via a package (so long as the package is higher than 0.3.13)
# It checks out the dokku source code from Github into ~/dokku and then runs 'make install' from dokku source.

# We wrap this whole script in a function, so that we won't execute
# until the entire script is downloaded.
# That's good because it prevents our output overlapping with wget's.
# It also means that we can't run a partially downloaded script.


bootstrap () {


set -eo pipefail
export DEBIAN_FRONTEND=noninteractive
export DOKKU_REPO=${DOKKU_REPO:-"https://github.com/dokku/dokku.git"}

echo "Preparing to install $DOKKU_TAG from $DOKKU_REPO..."
if ! command -v apt-get &>/dev/null; then
  echo "This installation script requires apt-get. For manual installation instructions, consult http://dokku.viewdocs.io/dokku/advanced-installation/"
  exit 1
fi

hostname -f > /dev/null 2>&1 || {
  echo "This installation script requires that you have a hostname set for the instance. Please set a hostname for 127.0.0.1 in your /etc/hosts"
  exit 1
}

echo "--> Ensuring we have the proper dependencies"
apt-get update -qq > /dev/null
[[ $(lsb_release -sr) == "12.04" ]] && apt-get install -qq -y python-software-properties

dokku_install_source() {
  apt-get install -qq -y git make software-properties-common
  cd /root
  if [[ ! -d /root/dokku ]]; then
    git clone $DOKKU_REPO /root/dokku
  fi

  cd /root/dokku
  git fetch origin
  git checkout $DOKKU_CHECKOUT
  make install
}

dokku_install_package() {
  echo "--> Initial apt-get update"
  apt-get update -qq > /dev/null
  apt-get install -qq -y apt-transport-https

  echo "--> Installing docker"
  if uname -r | grep -q linode; then
    echo "--> NOTE: Using Linode? Docker might complain about missing AUFS support."
    echo "    See http://dokku.viewdocs.io/dokku/getting-started/install/linode/"
    echo "    Installation will continue in 10 seconds."
    sleep 10
  fi
  wget -nv -O - https://get.docker.com/ | sh

  echo "--> Installing dokku"
  wget -nv -O - https://packagecloud.io/gpg.key | apt-key add -
  echo "deb https://packagecloud.io/dokku/dokku/ubuntu/ trusty main" | tee /etc/apt/sources.list.d/dokku.list
  apt-get update -qq > /dev/null

  if [[ -n $DOKKU_CHECKOUT ]]; then
    apt-get install -y dokku=$DOKKU_CHECKOUT
  else
    apt-get install -y dokku
  fi
}

if [[ -n $DOKKU_BRANCH ]]; then
  export DOKKU_CHECKOUT="origin/$DOKKU_BRANCH"
  dokku_install_source
elif [[ -n $DOKKU_TAG ]]; then
  export DOKKU_SEMVER="${DOKKU_TAG//v}"
  major=$(echo $DOKKU_SEMVER | awk '{split($0,a,"."); print a[1]}')
  minor=$(echo $DOKKU_SEMVER | awk '{split($0,a,"."); print a[2]}')
  patch=$(echo $DOKKU_SEMVER | awk '{split($0,a,"."); print a[3]}')

  # 0.3.13 was the first version with a debian package
  if [[ "$major" -eq "0" ]] && [[ "$minor" -eq "3" ]] && [[ "$patch" -ge "13" ]]; then
    export DOKKU_CHECKOUT="$DOKKU_SEMVER"
    dokku_install_package
    echo "--> Running post-install dependency installation"
    dokku plugins-install-dependencies
  # 0.4.0 implemented a `plugin` plugin
  elif [[ "$major" -eq "0" ]] && [[ "$minor" -ge "4" ]] && [[ "$patch" -ge "0" ]]; then
    export DOKKU_CHECKOUT="$DOKKU_SEMVER"
    dokku_install_package
    echo "--> Running post-install dependency installation"
    sudo -E dokku plugin:install-dependencies --core
  else
    export DOKKU_CHECKOUT="$DOKKU_TAG"
    dokku_install_source
  fi
else
  dokku_install_package
  echo "--> Running post-install dependency installation"
  sudo -E dokku plugin:install-dependencies --core
fi

}

bootstrap
