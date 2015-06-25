#!/usr/bin/env bash
set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

MOVES=(ABLE ABNORMA AGAIN AIREXPL ANG ANGER ASAIL ATTACK AURORA AWL BAN BAND BARE BEAT BEATED BELLY BIND BITE BLOC BLOOD BODY BOOK BREATH BUMP CAST CHAM CLAMP CLAP CLAW CLEAR CLI CLIP CLOUD CONTRO CONVY COOLHIT CRASH CRY CUT DESCRI D-FIGHT DIG DITCH DIV DOZ DRE DUL DU-PIN DYE EARTH EDU EG-BOMB EGG ELEGY ELE-HIT EMBODY EMPLI ENGL ERUPT EVENS EXPLOR EYES FALL FAST F-CAR F-DANCE FEARS F-FIGHT FIGHT FIR FIRE FIREHIT FLAME FLAP FLASH FLEW FORCE FRA FREEZE FROG G-BIRD GENKISS GIFT G-KISS G-MOUSE GRADE GROW HAMMER HARD HAT HATE H-BOMB HELL-R HEMP HINT HIT HU HUNT HYPNOSI INHA IRO IRONBAR IR-WING J-GUN KEE KICK KNIF KNIFE KNOCK LEVEL LIGH LIGHHIT LIGHT LIVE L-WALL MAD MAJUS MEL MELO MESS MILK MIMI MISS MIXING MOVE MUD NI-BED NOISY NOONLI NULL N-WAVE PAT PEACE PIN PLAN PLANE POIS POL POWDE POWE POWER PRIZE PROTECT PROUD RAGE RECOR REFLAC REFREC REGR RELIV RENEW R-FIGHT RING RKICK ROCK ROUND RUS RUSH SAND SAW SCISSOR SCRA SCRIPT SEEN SERVER SHADOW SHELL SHINE SHO SIGHT SIN SMALL SMELT SMOK SNAKE SNO SNOW SOU SO-WAVE SPAR SPEC SPID S-PIN SPRA STAM STARE STEA STONE STORM STRU STRUG STUDEN SUBS SUCID SUN-LIG SUNRIS SUPLY S-WAVE TAILS TANGL TASTE TELLI THANK TONKICK TOOTH TORL TRAIN TRIKICK TUNGE VOLT WA-GUN WATCH WAVE W-BOMB WFALL WFING WHIP WHIRL WIND WOLF WOOD WOR YUJA)
NAMES=(SEED GRASS FLOWE SHAD CABR SNAKE GOLD COW GUIKI PEDAL DELAN B-FLY BIDE KEYU FORK LAP PIGE PIJIA CAML LAT BIRD BABOO VIV ABOKE PIKAQ RYE SAN BREAD LIDEL LIDE PIP PIKEX ROK JUGEN PUD BUDE ZHIB GELU GRAS FLOW LAFUL ATH BALA CORN MOLUF DESP DAKED MIMI BOLUX KODA GELUD MONK SUMOY GEDI WENDI NILEM NILE NILEC KEZI YONGL HUDE WANLI GELI GUAIL MADAQ WUCI WUCI MUJEF JELLY SICIB GELU NELUO BOLI JIALE YED YEDE CLO SCARE AOCO DEDE DEDEI BAWU JIUG BADEB BADEB HOLE BALUX GES FANT QUAR YIHE SWAB SLIPP CLU DEPOS BILIY YUANO SOME NO YELA EMPT ZECUN XIAHE BOLEL DEJI MACID XIHON XITO LUCK MENJI GELU DECI XIDE DASAJ DONGN RICUL MINXI BALIY ZENDA LUZEL HELE5 0FENB KAIL JIAND CARP JINDE LAPU MUDE YIFU LINLI SANDI HUSI JINC OUMU OUMUX CAP KUIZA PUD TIAO FRMAN CLAU SPARK DRAGO BOLIU GUAIL MIYOU MIY QIAOK BEIL MUKEI RIDED MADAM BAGEP CROC ALIGE OUDAL OUD DADA HEHE YEDEA NUXI NUXIN ROUY ALIAD STICK QIANG LAAND PIQI PI PUPI DEKE DEKEJ NADI NADIO MALI PEA ELECT FLOWE MAL MALI HUSHU NILEE YUZI POPOZ DUZI HEBA XIAN SHAN YEYEA WUY LUO KEFE HULA CROW YADEH MOW ANNAN SUONI KYLI HULU HUDEL YEHE GULAE YEHE BLU GELAN BOAT NIP POIT HELAK XINL BEAR LINB MAGEH MAGEJ WULI YIDE RIVE FISH AOGU DELIE MANTE KONMU DELU HELU HUAN HUMA DONGF JINCA HEDE DEFU LIBY JIAPA MEJI HELE BUHU MILK HABI THUN GARD DON YANGQ SANAQ BANQ LUJ PHIX SIEI EGG)

random_number() {
  [[ -n "$1" ]] && RANGE="$1"
  if [[ -n "$RANGE" ]];then
    number=$RANDOM
    let "number %= $RANGE"
  else
    number=$RANDOM
  fi
  echo $number
}

random_name() {
 NUM1=$(random_number ${#MOVES[@]})
 NUM2=$(random_number ${#MOVES[@]})
 NUM3=$(random_number ${#NAMES[@]})

 UPPER_APPNAME="${MOVES[${NUM1}]}-${MOVES[${NUM2}]}-${NAMES[${NUM3}]}"

 [[ "$BASH_VERSION" =~ 4.* ]] && lower_appname=${UPPER_APPNAME,,}
 [[ -z "$lower_appname" ]] && lower_appname=$(echo $UPPER_APPNAME | tr '[:upper:]' '[:lower:]')
 echo "$lower_appname"
}

client_help_msg() {
  echo "==> Configure the DOKKU_HOST environment variable or run $0 from a repository with a git remote named dokku"
  echo "--> i.e. git remote add dokku dokku@<dokku-host>:<app-name>"
  exit 20 # exit with specific status. only used in units tests for now
}

if [[ -z $DOKKU_HOST ]]; then
  if [[ -d .git ]] || git rev-parse --git-dir > /dev/null 2>&1; then
    DOKKU_HOST=$(git remote -v 2>/dev/null | grep -Ei "^dokku" | head -n 1 | cut -f1 -d' ' | cut -f2 -d '@' | cut -f1 -d':' 2>/dev/null || true)
  else
    client_help_msg
  fi
fi

if [[ ! -z $DOKKU_HOST ]]; then
  _dokku() {
    appname=""
    if [[ -d .git ]] || git rev-parse --git-dir > /dev/null 2>&1; then
      set +e
      appname=$(git remote -v 2>/dev/null | grep -Ei "dokku@$DOKKU_HOST" | head -n 1 | cut -f1 -d' ' | cut -f2 -d':' 2>/dev/null)
      set -e
    else
      echo "This is not a git repository"
    fi

    if [[ "$appname" != "" ]] && [[ -n "$*" ]]; then
      case "$1" in
        apps|backup*|help|plugins*|ps:restartall|trace|version)
          true
          ;;
        apps:destroy)
          if [[ -z "$2" ]] || [[ "$2" == "force" ]];then
            donotshift="YES"
          fi
          ;;
        *)
          donotshift="YES"
          ;;
      esac
    fi

    if [[ "$1" = "apps:create" ]] && [[ -z "$2" ]]; then
      appname=$(random_name)
      counter=0
      while ssh "dokku@$DOKKU_HOST" apps 2>/dev/null| grep -q "$appname"; do
        if [[ $counter -ge 100 ]];then
          echo "Error: could not reasonably generate a new app name. try cleaning up some apps..."
          ssh "dokku@$DOKKU_HOST" apps
          exit 1
        else
          appname=$(random_name)
          counter=$((counter+1))
        fi
      done
      if git remote add dokku "dokku@$DOKKU_HOST:$appname"
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
      ssh -t "dokku@$DOKKU_HOST" "$@"
      exit $?
    fi

    if [[ -z "$verb" ]]; then
      if [[ ! "$1" =~ --* ]]; then
        verb=$1
        shift
        
        if [[ "$1" == "$appname" ]]; then
          shift
        fi
        
        args="$*"
      else
        long_args="--"
        for arg in "$@";do
          if [[ "$arg" =~ --* ]];then
            long_args+=" $arg"
            args=("${args[@]/$arg}")
          else
             verb="$arg"
          fi
        done
      fi
    fi
    # shellcheck disable=SC2086
    ssh -t "dokku@$DOKKU_HOST" $long_args "$verb" "$appname" "${args[@]}"
  }

  if [[ "$0" == "dokku" ]] || [[ "$0" == *dokku_client.sh ]] || [[ "$0" == $(which dokku) ]]; then
    _dokku "$@"
    exit $?
  fi
else
  client_help_msg
fi
