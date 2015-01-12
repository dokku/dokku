#!/usr/bin/env bash

if [[ ! -z $DOKKU_HOST ]]; then
	function dokku {
		appname=$(git remote -v 2>/dev/null | grep dokku | head -n 1 | cut -f1 -d' ' | cut -f2 -d':' 2>/dev/null)
		if [[ "$?" != "0" ]]; then
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
			ssh dokku@$DOKKU_HOST $*
			exit
		fi

		verb=$1
		shift
		ssh dokku@$DOKKU_HOST "$verb" "$appname" $@
	}
fi
