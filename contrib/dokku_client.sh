#!/usr/bin/env bash
set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

if [[ -z $DOKKU_HOST ]]; then
  if [[ -d .git ]] || git rev-parse --git-dir > /dev/null 2>&1; then
    DOKKU_HOST=$(git remote -v 2>/dev/null | grep -Ei "^dokku" | head -n 1 | cut -f1 -d' ' | cut -f2 -d '@' | cut -f1 -d':' 2>/dev/null)
  fi
fi

if [[ ! -z $DOKKU_HOST ]]; then
  function _dokku {
    appname=""
    if [[ -d .git ]] || git rev-parse --git-dir > /dev/null 2>&1; then
      set +e
      appname=$(git remote -v 2>/dev/null | grep -Ei "^dokku" | head -n 1 | cut -f1 -d' ' | cut -f2 -d':' 2>/dev/null)
      set -e
    else
      echo "This is not a git repository"
      exit 1
    fi

    if [[ "$appname" != "" ]]; then
      donotshift="YES"
    fi

    if [[ "$1" = "create" ]]; then
      appname=$(echo "print(elfs.GenName())" | lua -l elfs)
      if git remote add dokku dokku@$DOKKU_HOST:$appname
      then
        echo "-----> Dokku remote added at $DOKKU_HOST"
        echo "-----> Application name is $appname"
      else
        echo "!      Dokku remote not added! Do you already have a dokku remote?"
        return
      fi
      git push dokku master
      return $?
    fi

    if [[ -z "$donotshift" ]]; then
      ssh dokku@$DOKKU_HOST "$*"
      exit $?
    fi

    verb=$1
    shift
    ssh dokku@$DOKKU_HOST "$verb" "$appname" "$@"
  }

  if [[ "$0" == "dokku" ]] || [[ "$0" == *dokku_client.sh ]]; then
    _dokku "$@"
    exit $?
  fi
fi
