#!/usr/bin/env bash
set -eo pipefail; [[ $TRACE ]] && set -x

# A script to bootstrap dokku.
# It expects to be run on Ubuntu 14.04 via 'sudo'
# If installing a tag higher than 0.3.13, it may install dokku via a package (so long as the package is higher than 0.3.13)
# It checks out the dokku source code from Github into ~/dokku and then runs 'make install' from dokku source.

# We wrap this whole script in functions, so that we won't execute
# until the entire script is downloaded.
# That's good because it prevents our output overlapping with wget's.
# It also means that we can't run a partially downloaded script.

ensure-environment() {
  local FREE_MEMORY
  echo "Preparing to install $DOKKU_TAG from $DOKKU_REPO..."
  if ! command -v apt-get &>/dev/null; then
    echo "This installation script requires apt-get. For manual installation instructions, consult http://dokku.viewdocs.io/dokku/advanced-installation/"
    exit 1
  fi

  hostname -f > /dev/null 2>&1 || {
    echo "This installation script requires that you have a hostname set for the instance. Please set a hostname for 127.0.0.1 in your /etc/hosts"
    exit 1
  }

  FREE_MEMORY=$(grep MemTotal /proc/meminfo | awk '{print $2}')
  if [[ "$FREE_MEMORY" -lt 1003600 ]]; then
    echo "For dokku to build containers, it is strongly suggested that you have 1024 megabytes or more of free memory"
    echo "If necessary, please consult this document to setup swap: http://dokku.viewdocs.io/dokku/advanced-installation/#vms-with-less-than-1gb-of-memory"
  fi
}

install-requirements() {
  echo "--> Ensuring we have the proper dependencies"
  apt-get update -qq > /dev/null
  if [[ $(lsb_release -sr) == "12.04" ]]; then
    apt-get -qq -y install python-software-properties
  fi
}

install-dokku() {
  if [[ -n $DOKKU_BRANCH ]]; then
    install-dokku-from-source "origin/$DOKKU_BRANCH"
  elif [[ -n $DOKKU_TAG ]]; then
    local DOKKU_SEMVER="${DOKKU_TAG//v}"
    major=$(echo "$DOKKU_SEMVER" | awk '{split($0,a,"."); print a[1]}')
    minor=$(echo "$DOKKU_SEMVER" | awk '{split($0,a,"."); print a[2]}')
    patch=$(echo "$DOKKU_SEMVER" | awk '{split($0,a,"."); print a[3]}')

    # 0.3.13 was the first version with a debian package
    if [[ "$major" -eq "0" ]] && [[ "$minor" -eq "3" ]] && [[ "$patch" -ge "13" ]]; then
      install-dokku-from-package "$DOKKU_SEMVER"
      echo "--> Running post-install dependency installation"
      dokku plugins-install-dependencies
    # 0.4.0 implemented a `plugin` plugin
    elif [[ "$major" -eq "0" ]] && [[ "$minor" -ge "4" ]] && [[ "$patch" -ge "0" ]]; then
      install-dokku-from-package "$DOKKU_SEMVER"
      echo "--> Running post-install dependency installation"
      sudo -E dokku plugin:install-dependencies --core
    else
      install-dokku-from-source "$DOKKU_TAG"
    fi
  else
    install-dokku-from-package
    echo "--> Running post-install dependency installation"
    sudo -E dokku plugin:install-dependencies --core
  fi
}


install-dokku-from-source() {
  local DOKKU_CHECKOUT="$1"
  apt-get -qq -y install git make software-properties-common
  cd /root
  if [[ ! -d /root/dokku ]]; then
    git clone "$DOKKU_REPO" /root/dokku
  fi

  cd /root/dokku
  git fetch origin
  [[ -n $DOKKU_CHECKOUT ]] && git checkout "$DOKKU_CHECKOUT"
  make install
}

install-dokku-from-package() {
  local DOKKU_CHECKOUT="$1"
  local NO_INSTALL_RECOMMENDS=${DOKKU_NO_INSTALL_RECOMMENDS:=""}

  if [[ -n $DOKKU_DOCKERFILE ]]; then
    NO_INSTALL_RECOMMENDS=" --no-install-recommends "
  fi

  echo "--> Initial apt-get update"
  apt-get update -qq > /dev/null
  apt-get -qq -y install apt-transport-https

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

  [[ -n $DOKKU_VHOST_ENABLE ]]  && echo "dokku dokku/vhost_enable boolean $DOKKU_VHOST_ENABLE"   | sudo debconf-set-selections
  [[ -n $DOKKU_WEB_CONFIG ]]    && echo "dokku dokku/web_config boolean $DOKKU_WEB_CONFIG"       | sudo debconf-set-selections
  [[ -n $DOKKU_HOSTNAME ]]      && echo "dokku dokku/hostname string $DOKKU_HOSTNAME"            | sudo debconf-set-selections
  [[ -n $DOKKU_SKIP_KEY_FILE ]] && echo "dokku dokku/skip_key_file boolean $DOKKU_SKIP_KEY_FILE" | sudo debconf-set-selections
  [[ -n $DOKKU_KEY_FILE ]]      && echo "dokku dokku/key_file string $DOKKU_KEY_FILE"            | sudo debconf-set-selections

  if [[ -n $DOKKU_CHECKOUT ]]; then
    # shellcheck disable=SC2086
    apt-get -qq -y $NO_INSTALL_RECOMMENDS install "dokku=$DOKKU_CHECKOUT"
  else
    # shellcheck disable=SC2086
    apt-get -qq -y $NO_INSTALL_RECOMMENDS install dokku
  fi
}

main() {
  export DEBIAN_FRONTEND=noninteractive
  export DOKKU_REPO=${DOKKU_REPO:-"https://github.com/dokku/dokku.git"}

  ensure-environment
  install-requirements
  install-dokku
}

main "$@"
