#!/usr/bin/env bash
set -eo pipefail
[[ $DOKKU_TRACE ]] && set -x
export DOKKU_PORT=${DOKKU_PORT:=22}
export DOKKU_HOST=${DOKKU_HOST:=}

fn-random-number() {
  [[ -n "$1" ]] && RANGE="$1"
  if [[ -n "$RANGE" ]]; then
    number=$RANDOM
    let "number %= $RANGE"
  else
    number=$RANDOM
  fi
  echo $number
}

fn-random-name() {
  local NUM1 NUM2 NUM3 UPPER_APPNAME lower_appname
  local MOVES=(ABLE ABNORMA AGAIN AIREXPL ANG ANGER ASAIL ATTACK AURORA AWL BAN BAND BARE BEAT BEATED BELLY BIND BITE BLOC BLOOD BODY BOOK BREATH BUMP CAST CHAM CLAMP CLAP CLAW CLEAR CLI CLIP CLOUD CONTRO CONVY COOLHIT CRASH CRY CUT DESCRI D-FIGHT DIG DITCH DIV DOZ DRE DUL DU-PIN DYE EARTH EDU EG-BOMB EGG ELEGY ELE-HIT EMBODY EMPLI ENGL ERUPT EVENS EXPLOR EYES FALL FAST F-CAR F-DANCE FEARS F-FIGHT FIGHT FIR FIRE FIREHIT FLAME FLAP FLASH FLEW FORCE FRA FREEZE FROG G-BIRD GENKISS GIFT G-KISS G-MOUSE GRADE GROW HAMMER HARD HAT HATE H-BOMB HELL-R HEMP HINT HIT HU HUNT HYPNOSI INHA IRO IRONBAR IR-WING J-GUN KEE KICK KNIF KNIFE KNOCK LEVEL LIGH LIGHHIT LIGHT LIVE L-WALL MAD MAJUS MEL MELO MESS MILK MIMI MISS MIXING MOVE MUD NI-BED NOISY NOONLI NULL N-WAVE PAT PEACE PIN PLAN PLANE POIS POL POWDE POWE POWER PRIZE PROTECT PROUD RAGE RECOR REFLAC REFREC REGR RELIV RENEW R-FIGHT RING RKICK ROCK ROUND RUS RUSH SAND SAW SCISSOR SCRA SCRIPT SEEN SERVER SHADOW SHELL SHINE SHO SIGHT SIN SMALL SMELT SMOK SNAKE SNO SNOW SOU SO-WAVE SPAR SPEC SPID S-PIN SPRA STAM STARE STEA STONE STORM STRU STRUG STUDEN SUBS SUCID SUN-LIG SUNRIS SUPLY S-WAVE TAILS TANGL TASTE TELLI THANK TONKICK TOOTH TORL TRAIN TRIKICK TUNGE VOLT WA-GUN WATCH WAVE W-BOMB WFALL WFING WHIP WHIRL WIND WOLF WOOD WOR YUJA)
  local NAMES=(SEED GRASS FLOWE SHAD CABR SNAKE GOLD COW GUIKI PEDAL DELAN B-FLY BIDE KEYU FORK LAP PIGE PIJIA CAML LAT BIRD BABOO VIV ABOKE PIKAQ RYE SAN BREAD LIDEL LIDE PIP PIKEX ROK JUGEN PUD BUDE ZHIB GELU GRAS FLOW LAFUL ATH BALA CORN MOLUF DESP DAKED MIMI BOLUX KODA GELUD MONK SUMOY GEDI WENDI NILEM NILE NILEC KEZI YONGL HUDE WANLI GELI GUAIL MADAQ WUCI WUCI MUJEF JELLY SICIB GELU NELUO BOLI JIALE YED YEDE CLO SCARE AOCO DEDE DEDEI BAWU JIUG BADEB BADEB HOLE BALUX GES FANT QUAR YIHE SWAB SLIPP CLU DEPOS BILIY YUANO SOME NO YELA EMPT ZECUN XIAHE BOLEL DEJI MACID XIHON XITO LUCK MENJI GELU DECI XIDE DASAJ DONGN RICUL MINXI BALIY ZENDA LUZEL HELE5 0FENB KAIL JIAND CARP JINDE LAPU MUDE YIFU LINLI SANDI HUSI JINC OUMU OUMUX CAP KUIZA PUD TIAO FRMAN CLAU SPARK DRAGO BOLIU GUAIL MIYOU MIY QIAOK BEIL MUKEI RIDED MADAM BAGEP CROC ALIGE OUDAL OUD DADA HEHE YEDEA NUXI NUXIN ROUY ALIAD STICK QIANG LAAND PIQI PI PUPI DEKE DEKEJ NADI NADIO MALI PEA ELECT FLOWE MAL MALI HUSHU NILEE YUZI POPOZ DUZI HEBA XIAN SHAN YEYEA WUY LUO KEFE HULA CROW YADEH MOW ANNAN SUONI KYLI HULU HUDEL YEHE GULAE YEHE BLU GELAN BOAT NIP POIT HELAK XINL BEAR LINB MAGEH MAGEJ WULI YIDE RIVE FISH AOGU DELIE MANTE KONMU DELU HELU HUAN HUMA DONGF JINCA HEDE DEFU LIBY JIAPA MEJI HELE BUHU MILK HABI THUN GARD DON YANGQ SANAQ BANQ LUJ PHIX SIEI EGG)

  NUM1=$(fn-random-number ${#MOVES[@]})
  NUM2=$(fn-random-number ${#MOVES[@]})
  NUM3=$(fn-random-number ${#NAMES[@]})

  UPPER_APPNAME="${MOVES[${NUM1}]}-${MOVES[${NUM2}]}-${NAMES[${NUM3}]}"

  [[ "$BASH_VERSION" =~ 4.* ]] && lower_appname=${UPPER_APPNAME,,}
  [[ -z "$lower_appname" ]] && lower_appname=$(echo "$UPPER_APPNAME" | tr '[:upper:]' '[:lower:]')
  echo "$lower_appname"
}

fn-client-help-msg() {
  echo "=====> Configure the DOKKU_HOST environment variable or run $0 from a repository with a git remote named dokku"
  echo "       i.e. git remote add dokku dokku@<dokku-host>:<app-name>"
  exit 20 # exit with specific status. only used in units tests for now
}

fn-is-git-repo() {
  git rev-parse &>/dev/null
}

fn-has-dokku-remote() {
  git remote show | grep -E "^${DOKKU_GIT_REMOTE}\s"
}

fn-dokku-host() {
  declare DOKKU_GIT_REMOTE="$1" DOKKU_HOST="$2"

  if [[ -z "$DOKKU_HOST" ]]; then
    if [[ -d .git ]] || git rev-parse --git-dir >/dev/null 2>&1; then
      DOKKU_HOST=$(git remote -v 2>/dev/null | grep -Ei "^${DOKKU_GIT_REMOTE}\s" | head -n 1 | cut -f1 -d' ' | cut -f2 -d '@' | cut -f1 -d':' 2>/dev/null || true)
    fi
  fi

  if [[ -z "$DOKKU_HOST" ]]; then
    return
  fi

  echo "$DOKKU_HOST"
}

fn-get-remote() {
  git config dokku.remote 2>/dev/null || echo "dokku"
}

main() {
  declare CMD="$1" APP_ARG="$2"
  local APP="" DOKKU_GIT_REMOTE="$(fn-get-remote)" DOKKU_REMOTE_HOST=""
  local cmd_set=false next_index=1 skip=false args=("$@")

  for arg in "$@"; do
    if [[ "$skip" == "true" ]]; then
      next_index=$((next_index + 1))
      skip=false
      continue
    fi
    is_flag=false

    [[ "$arg" =~ ^--.* ]] && is_flag=true

    if [[ "$arg" == "--app" ]]; then
      APP=${args[$next_index]}
      skip=true
      shift 2
    elif [[ "$arg" == "--remote" ]]; then
      DOKKU_GIT_REMOTE=${args[$next_index]}
      skip=true
      shift 2
    elif [[ "$arg" =~ ^--.* ]]; then
      [[ "$cmd_set" == "true" ]] && [[ "$is_flag" == "false" ]] && APP_ARG="$arg" && break
      [[ "$arg" == "--trace" ]] && export DOKKU_TRACE=1 && set -x
    else
      if [[ "$cmd_set" == "true" ]] && [[ "$is_flag" == "false" ]]; then
        APP_ARG="$arg"
        break
      else
        CMD="$arg"
        cmd_set=true
      fi
    fi
    next_index=$((next_index + 1))
  done

  DOKKU_REMOTE_HOST="$(fn-dokku-host "$DOKKU_GIT_REMOTE" "$DOKKU_HOST")"
  if [[ -z "$DOKKU_REMOTE_HOST" ]] && [[ "$CMD" != remote ]] && [[ "$CMD" != remote:* ]]; then
    fn-client-help-msg
  fi

  if [[ -z "$APP" ]]; then
    if [[ -d .git ]] || git rev-parse --git-dir >/dev/null 2>&1; then
      set +e
      APP=$(git remote -v 2>/dev/null | grep -Ei "^${DOKKU_GIT_REMOTE}\s" | grep -Ei "dokku@$DOKKU_REMOTE_HOST" | head -n 1 | cut -f2 -d'@' | cut -f1 -d' ' | cut -f2 -d':' 2>/dev/null)
      set -e
    else
      echo " !     This is not a git repository" 1>&2
    fi
  fi

  case "$CMD" in
    apps:create)
      if [[ -z "$APP" ]] && [[ -z "$APP_ARG" ]]; then
        APP=$(fn-random-name)
        counter=0
        while ssh -p "$DOKKU_PORT" "dokku@$DOKKU_REMOTE_HOST" apps 2>/dev/null | grep -q "$APP"; do
          if [[ $counter -ge 100 ]]; then
            echo " !     Could not reasonably generate a new app name. Try cleaning up some apps..." 1>&2
            ssh -p "$DOKKU_PORT" "dokku@$DOKKU_REMOTE_HOST" apps
            exit 1
          else
            APP=$(random_name)
            counter=$((counter + 1))
          fi
        done
      elif [[ -z "$APP" ]]; then
        APP="$APP_ARG"
      fi
      if git remote add "$DOKKU_GIT_REMOTE" "dokku@$DOKKU_REMOTE_HOST:$APP"; then
        echo "-----> Dokku remote added at ${DOKKU_REMOTE_HOST} called ${DOKKU_GIT_REMOTE}"
        echo "-----> Application name is ${APP}"
      else
        echo " !     Dokku remote not added! Do you already have a dokku remote?" 1>&2
        return
      fi
      ;;
    apps:destroy)
      fn-is-git-repo && fn-has-dokku-remote && git remote remove "$DOKKU_GIT_REMOTE"
      ;;
    remote)
      echo "$DOKKU_GIT_REMOTE"
      exit 0
      ;;
    remote:add)
      shift 1
      git remote add "$@"
      exit "$?"
      ;;
    remote:list)
      git remote
      exit "$?"
      ;;
    remote:set)
      shift 1
      git config dokku.remote "$@"
      exit "$?"
      ;;
    remote:remove)
      shift 1
      git remote remove "$@"
      exit "$?"
      ;;
    remote:unset)
      git config --unset dokku.remote
      exit "$?"
      ;;
  esac

  [[ " apps certs help ls nginx shell storage trace version " == *" $CMD "* ]] && unset APP
  [[ " certs:chain domains:add-global domains:remove-global domains:set-global ps:restore " == *" $CMD "* ]] && unset APP
  [[ " storage:ensure-directory " == *" $CMD "* ]] && unset APP
  [[ "$CMD" =~ events*|plugin*|ssh-keys* ]] && unset APP
  [[ -n "$APP_ARG" ]] && [[ "$APP_ARG" == "--global" ]] && unset APP
  [[ -n "$@" ]] && [[ -n "$APP" ]] && app_arg="--app $APP"
  # echo "ssh -o LogLevel=QUIET -p $DOKKU_PORT -t dokku@$DOKKU_REMOTE_HOST -- $app_arg $@"
  # shellcheck disable=SC2068,SC2086
  ssh -o LogLevel=QUIET -p $DOKKU_PORT -t dokku@$DOKKU_REMOTE_HOST -- $app_arg $@ || {
    ssh_exit_code="$?"
    echo " !     Failed to execute dokku command over ssh: exit code $?" 1>&2
    echo " !     If there was no output from Dokku, ensure your configured SSH Key can connect to the remote server" 1>&2
    return $ssh_exit_code
  }
}

if [[ "$0" == "dokku" ]] || [[ "$0" == *dokku_client.sh ]] || [[ "$0" == $(command -v dokku) ]]; then
  main "$@"
  exit $?
fi
