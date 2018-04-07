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

  case "$DOKKU_DISTRO" in
    debian|ubuntu)
      apt-get update -qq > /dev/null
      ;;
  esac
}

install-dokku() {
  if [[ -n $DOKKU_BRANCH ]]; then
    install-dokku-from-source "origin/$DOKKU_BRANCH"
  elif [[ -n $DOKKU_TAG ]]; then
    local DOKKU_SEMVER="${DOKKU_TAG//v}"
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
    echo "This installation script requires apt-get. For manual installation instructions, consult http://dokku.viewdocs.io/dokku/advanced-installation/"
    exit 1
  fi

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
  case "$DOKKU_DISTRO" in
    debian|ubuntu)
      install-dokku-from-deb-package "$@"
      ;;
    centos|rhel)
      install-dokku-from-rpm-package "$@"
      ;;
    *)
      echo "Unsupported Linux distribution. For manual installation instructions, consult http://dokku.viewdocs.io/dokku/advanced-installation/"
      exit 1
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

  if [[ -n $DOKKU_DOCKERFILE ]]; then
    NO_INSTALL_RECOMMENDS=" --no-install-recommends "
  fi

  echo "--> Initial apt-get update"
  apt-get update -qq > /dev/null
  apt-get -qq -y install apt-transport-https

  echo "--> Installing docker"
  if uname -r | grep -q linode; then
    echo "--> NOTE: Using Linode? Docker may complain about missing AUFS support."
    echo "    You can safely ignore this warning."
    echo "    Installation will continue in 10 seconds."
    sleep 10
  fi
  wget -nv -O - https://get.docker.com/ | sh

  if [[ "$DOKKU_DISTRO_VERSION" == "14.04" ]]; then
    echo "--> Adding nginx PPA"
    add-apt-repository -y ppa:nginx/stable
  fi

  OS_ID="$(lsb_release -cs 2> /dev/null || echo "xenial")"
  if ! in-array "$DOKKU_DISTRO" "debian" "ubuntu"; then
    DOKKU_DISTRO="ubuntu"
    OS_ID="xenial"
  fi

  if [[ "$DOKKU_DISTRO" == "ubuntu" ]]; then
    OS_IDS=("trusty" "utopic" "vivid" "wily" "xenial" "yakkety" "zesty" "artful" "bionic")
    if ! in-array "$OS_ID" "${OS_IDS[@]}"; then
      OS_ID="xenial"
    fi
  elif [[ "$DOKKU_DISTRO" == "debian" ]]; then
    OS_IDS=("wheezy" "jessie" "stretch" "buster" "bullseye")
    if ! in-array "$OS_ID" "${OS_IDS[@]}"; then
      OS_ID="xenial"
    fi
  fi

  echo "--> Installing dokku"
  wget -nv -O - https://packagecloud.io/gpg.key | apt-key add -
  echo "deb https://packagecloud.io/dokku/dokku/$DOKKU_DISTRO/ $OS_ID main" | tee /etc/apt/sources.list.d/dokku.list
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

install-dokku-from-rpm-package() {
  local DOKKU_CHECKOUT="$1"

  if [[ "$DOKKU_DISTRO_VERSION" != "7" ]]; then
    echo "Only CentOS version 7 is supported."
    exit 1
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
