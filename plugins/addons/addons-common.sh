
ADDONS_PATH=/var/lib/dokku/addons

function die
{
  $QUIET || echo $* >&2
  exit 1
}

function export_addon_vars()
{
  export ADDON="$1"
  export ADDON_DATA="$DOKKU_ROOT/.addons/$ADDON"
  export ADDON_ROOT="$ADDONS_PATH/$ADDON"
}

function check_app
{
  if [[ -z $1 ]]; then
    die "You must specify an app name"
  else
    APP="$1"
    ADDONS_FILE="$DOKKU_ROOT/$APP/ADDONS"
    ENV_FILE="$DOKKU_ROOT/$APP/ENV"

    # Check if app exists with the same name
    if [ ! -d "$DOKKU_ROOT/$APP" ]; then
      die "App $APP does not exist"
    fi

    [ -f $ADDONS_FILE ] || {
      $QUIET || echo "-----> Creating $ADDONS_FILE"
      touch $ADDONS_FILE
    }
    [ -f $ENV_FILE ] || {
      $QUIET || echo "-----> Creating $ENV_FILE"
      touch $ENV_FILE
    }
  fi
}

function check_addon
{
  if [[ -z $1 ]]; then
    die "You must specify an addon name"
  elif [ ! -d "$ADDONS_PATH/$1" ]; then
    die "Addon $1 does not exist"
  fi
  export_addon_vars $1

  mkdir -p $ADDON_DATA
}

function check_addon_enabled
{
  check_addon $1
  if [ ! -f $ADDON_DATA/enabled ]; then
    die "Add-on $ADDON is not enabled"
  fi
}

function check_addon_disabled
{
  check_addon $1
  if [ -f $ADDON_DATA/enabled ]; then
    die "Add-on $ADDON is already enabled"
  fi
}

function check_addon_provisioned
{
  local line
  line=$(grep "^$ADDON;" $ADDONS_FILE) || {
    die "App $APP does not have addon $ADDON"
  }

  split_addon_line $line _ ADDON_ID ADDON_PRIVATE
}

function check_addon_unprovisioned
{
  if grep -q "^$ADDON;" $ADDONS_FILE; then
    die "App $APP already has addon $ADDON"
  fi
}

function split_addon_line
{
  parts=($(echo $1 | sed 's/;/ /g'))
  if [[ ! -z $2 && $2 != "_" ]]; then
    eval "$2=${parts[0]}"
  fi
  if [[ ! -z $3 && $3 != "_" ]]; then
    eval "$3=${parts[1]}"
  fi
  if [[ ! -z $4 && $4 != "_" ]]; then
    eval "$4=${parts[2]}"
  fi
}


