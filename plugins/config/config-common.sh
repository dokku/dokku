
if [[ -z $1 ]]; then
  echo "You must specify an app name"
  exit 1
else
  APP="$1"
  APP_DIR="$DOKKU_ROOT/$APP"
  
  ENV_FILE="$APP_DIR/ENV"

  # Check if app exists with the same name
  if [ ! -d "$APP_DIR" ]; then
    echo "App $APP does not exist"
    exit 1
  fi

  [ -f $ENV_FILE ] || {
    echo "-----> Creating $ENV_FILE"
    touch $ENV_FILE
  }
fi

config_styled_hash () {
  vars=`echo -e "$1"`

  longest=""
  for word in $vars; do
    KEY=`echo $word | cut -d"=" -f1`
    if [ ${#KEY} -gt ${#longest} ]; then
      longest=$KEY
    fi
  done

  for word in $vars; do
    KEY=`echo $word | cut -d"=" -f1`
    VALUE=`echo $word | cut -d"=" -f2-`

    num_zeros=$((${#longest} - ${#KEY}))
    zeros=" "
    while [ $num_zeros -gt 0 ]; do
      zeros="$zeros "
      num_zeros=$(($num_zeros - 1))
    done
    echo "$KEY:$zeros$VALUE"
  done
}

config_restart_app() {
  APP="$1"; IMAGE="app/$APP"

  echo "-----> Releasing $APP ..."
  dokku release $APP $IMAGE
  echo "-----> Release complete!"
  echo "-----> Deploying $APP ..."
  dokku deploy $APP $IMAGE
  echo "-----> Deploy complete!"
}

