#!/usr/bin/env bash
set -eo pipefail
[[ $TRACE ]] && set -x

# A script to bootstrap dokku.
# It expects to be run on Ubuntu 18.04/20.04, or CentOS 7 via 'sudo'
# If installing a tag higher than 0.3.13, it may install dokku via a package (so long as the package is higher than 0.3.13)
# It checks out the dokku source code from GitHub into ~/dokku and then runs 'make install' from dokku source.

# We wrap this whole script in functions, so that we won't execute
# until the entire script is downloaded.
# That's good because it prevents our output overlapping with wget's.
# It also means that we can't run a partially downloaded script.

SUPPORTED_VERSIONS="Debian [9, 10], CentOS [7], Fedora (partial) [33, 34], Ubuntu [18.04, 20.04]"

log-fail() {
  declare desc="log fail formatter"
  echo "$@" 1>&2
  exit 1
}

ensure-environment() {
  local FREE_MEMORY
  if [[ -z "$DOKKU_TAG" ]]; then
    echo "Preparing to install $DOKKU_REPO..."
  else
    echo "Preparing to install $DOKKU_TAG from $DOKKU_REPO..."
  fi

  hostname -f >/dev/null 2>&1 || {
    log-fail "This installation script requires that you have a hostname set for the instance. Please set a hostname for 127.0.0.1 in your /etc/hosts"
  }

  FREE_MEMORY=$(grep MemTotal /proc/meminfo | awk '{print $2}')
  if [[ "$FREE_MEMORY" -lt 1003600 ]]; then
    echo "For dokku to build containers, it is strongly suggested that you have 1024 megabytes or more of free memory"
    echo "If necessary, please consult this document to setup swap: https://dokku.com/docs/getting-started/advanced-installation/#vms-with-less-than-1-gb-of-memory"
  fi
}

install-requirements() {
  echo "--> Ensuring we have the proper dependencies"

  case "$DOKKU_DISTRO" in
    debian)
      if ! dpkg -l | grep -q gpg-agent; then
        apt-get update -qq >/dev/null
        apt-get -qq -y --no-install-recommends install gpg-agent
      fi
      if ! dpkg -l | grep -q software-properties-common; then
        apt-get update -qq >/dev/null
        apt-get -qq -y --no-install-recommends install software-properties-common
      fi
      ;;
    ubuntu)
      if ! dpkg -l | grep -q gpg-agent; then
        apt-get update -qq >/dev/null
        apt-get -qq -y --no-install-recommends install gpg-agent
      fi
      if ! dpkg -l | grep -q software-properties-common; then
        apt-get update -qq >/dev/null
        apt-get -qq -y --no-install-recommends install software-properties-common
      fi

      add-apt-repository universe >/dev/null
      apt-get update -qq >/dev/null
      ;;
  esac
}

install-dokku() {
  if ! command -v dokku &>/dev/null; then
    echo "--> Note: Installing dokku for the first time will result in removal of"
    echo "    files in the nginx 'sites-enabled' directory. Please manually"
    echo "    restore any files that may be removed after the installation and"
    echo "    web setup is complete."
    echo ""
    echo "    Installation will continue in 10 seconds."
    sleep 10
  fi

  if [[ -n $DOKKU_BRANCH ]]; then
    install-dokku-from-source "origin/$DOKKU_BRANCH"
  elif [[ -n $DOKKU_TAG ]]; then
    local DOKKU_SEMVER="${DOKKU_TAG//v/}"
    major=$(echo "$DOKKU_SEMVER" | awk '{split($0,a,"."); print a[1]}')
    minor=$(echo "$DOKKU_SEMVER" | awk '{split($0,a,"."); print a[2]}')
    patch=$(echo "$DOKKU_SEMVER" | awk '{split($0,a,"."); print a[3]}')

    use_plugin=false
    # 0.4.0 implemented a `plugin` plugin
    if [[ "$major" -eq "0" ]] && [[ "$minor" -ge "4" ]] && [[ "$patch" -ge "0" ]]; then
      use_plugin=true
    elif [[ "$major" -ge "1" ]]; then
      use_plugin=true
    fi

    # 0.3.13 was the first version with a debian package
    if [[ "$major" -eq "0" ]] && [[ "$minor" -eq "3" ]] && [[ "$patch" -ge "13" ]]; then
      install-dokku-from-package "$DOKKU_SEMVER"
      echo "--> Running post-install dependency installation"
      dokku plugins-install-dependencies
    elif [[ "$use_plugin" == "true" ]]; then
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

  if ! command -v apt-get &>/dev/null; then
    log-fail "This installation script requires apt-get. For manual installation instructions, consult https://dokku.com/docs/getting-started/advanced-installation/"
  fi

  apt-get -qq -y --no-install-recommends install sudo git make software-properties-common
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
  case "$DOKKU_DISTRO" in
    debian | ubuntu)
      install-dokku-from-deb-package "$@"
      ;;
    centos | fedora | rhel)
      install-dokku-from-rpm-package "$@"
      ;;
    *)
      log-fail "Unsupported Linux distribution. For manual installation instructions, consult https://dokku.com/docs/getting-started/advanced-installation/"
      ;;
  esac
}

in-array() {
  declare desc="return true if value ($1) is in list (all other arguments)"

  local e
  for e in "${@:2}"; do
    [[ "$e" == "$1" ]] && return 0
  done
  return 1
}

install-dokku-from-deb-package() {
  local DOKKU_CHECKOUT="$1"
  local NO_INSTALL_RECOMMENDS=${DOKKU_NO_INSTALL_RECOMMENDS:=""}
  local OS_ID

  if ! in-array "$DOKKU_DISTRO_VERSION" "18.04" "20.04" "9" "10"; then
    log-fail "Unsupported Linux distribution. Only the following versions are supported: $SUPPORTED_VERSIONS"
  fi

  if [[ -n $DOKKU_DOCKERFILE ]]; then
    NO_INSTALL_RECOMMENDS=" --no-install-recommends "
  fi

  echo "--> Initial apt-get update"
  apt-get update -qq >/dev/null
  apt-get -qq -y --no-install-recommends install apt-transport-https

  if ! command -v docker &>/dev/null; then
    echo "--> Installing docker"
    if uname -r | grep -q linode; then
      echo "--> NOTE: Using Linode? Docker may complain about missing AUFS support."
      echo "    You can safely ignore this warning."
      echo ""
      echo "    Installation will continue in 10 seconds."
      sleep 10
    fi
    export CHANNEL=stable
    wget -nv -O - https://get.docker.com/ | sh
  fi

  OS_ID="$(lsb_release -cs 2>/dev/null || echo "bionic")"
  if ! in-array "$DOKKU_DISTRO" "debian" "ubuntu"; then
    DOKKU_DISTRO="ubuntu"
    OS_ID="bionic"
  fi

  if [[ "$DOKKU_DISTRO" == "ubuntu" ]]; then
    OS_IDS=("bionic" "focal")
    if ! in-array "$OS_ID" "${OS_IDS[@]}"; then
      OS_ID="bionic"
    fi
  elif [[ "$DOKKU_DISTRO" == "debian" ]]; then
    OS_IDS=("stretch" "buster")
    if ! in-array "$OS_ID" "${OS_IDS[@]}"; then
      OS_ID="buster"
    fi
  fi

  echo "--> Installing dokku"
  wget -nv -O - https://packagecloud.io/dokku/dokku/gpgkey | apt-key add -
  echo "deb https://packagecloud.io/dokku/dokku/$DOKKU_DISTRO/ $OS_ID main" | tee /etc/apt/sources.list.d/dokku.list
  apt-get update -qq >/dev/null

  [[ -n $DOKKU_VHOST_ENABLE ]] && echo "dokku dokku/vhost_enable boolean $DOKKU_VHOST_ENABLE" | sudo debconf-set-selections
  [[ -n $DOKKU_WEB_CONFIG ]] && echo "dokku dokku/web_config boolean $DOKKU_WEB_CONFIG" | sudo debconf-set-selections
  [[ -n $DOKKU_HOSTNAME ]] && echo "dokku dokku/hostname string $DOKKU_HOSTNAME" | sudo debconf-set-selections
  [[ -n $DOKKU_SKIP_KEY_FILE ]] && echo "dokku dokku/skip_key_file boolean $DOKKU_SKIP_KEY_FILE" | sudo debconf-set-selections
  [[ -n $DOKKU_KEY_FILE ]] && echo "dokku dokku/key_file string $DOKKU_KEY_FILE" | sudo debconf-set-selections
  [[ -n $DOKKU_NGINX_ENABLE ]] && echo "dokku dokku/nginx_enable string $DOKKU_NGINX_ENABLE" | sudo debconf-set-selections

  if [[ -n $DOKKU_CHECKOUT ]]; then
    # shellcheck disable=SC2086
    apt-get -qq -y $NO_INSTALL_RECOMMENDS install "dokku=$DOKKU_CHECKOUT"
  else
    # shellcheck disable=SC2086
    apt-get -qq -y $NO_INSTALL_RECOMMENDS install dokku
  fi
}

install-dokku-from-rpm-package() {
  local DOKKU_CHECKOUT="$1"

  if ! in-array "$DOKKU_DISTRO_VERSION" "7"; then
    log-fail "Unsupported Linux distribution. Only the following versions are supported: $SUPPORTED_VERSIONS"
  fi

  echo "--> Installing docker"
  curl -fsSL https://get.docker.com/ | sh

  echo "--> Installing epel for nginx packages to be available"
  yum install -y epel-release

  echo "--> Installing herokuish and dokku"
  curl -s https://packagecloud.io/install/repositories/dokku/dokku/script.rpm.sh | bash
  if [[ -n $DOKKU_CHECKOUT ]]; then
    yum -y install herokuish "dokku-$DOKKU_CHECKOUT"
  else
    yum -y install herokuish dokku
  fi

  echo "--> Enabling docker and nginx on system startup"
  systemctl enable docker
  systemctl enable nginx

  echo "--> Starting nginx"
  systemctl start nginx
}

main() {
  export DOKKU_DISTRO DOKKU_DISTRO_VERSION
  # shellcheck disable=SC1091
  DOKKU_DISTRO=$(. /etc/os-release && echo "$ID")
  # shellcheck disable=SC1091
  DOKKU_DISTRO_VERSION=$(. /etc/os-release && echo "$VERSION_ID")

  export DEBIAN_FRONTEND=noninteractive
  export DOKKU_REPO=${DOKKU_REPO:-"https://github.com/dokku/dokku.git"}

  ensure-environment
  install-requirements
  install-dokku
}

main "$@"
