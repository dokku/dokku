#!/usr/bin/env bash
# shellcheck disable=SC2329
set -eo pipefail
[[ $TRACE ]] && set -x

# A script to bootstrap dokku.
# It expects to be run on Ubuntu 22.04/24.04/26.04 or Debian 11/12/13 via 'sudo'
# If installing a tag higher than 0.3.13, it may install dokku via a package (so long as the package is higher than 0.3.13)
# Otherwise it checks out the dokku source code from GitHub into ~/dokku and then runs 'make install' from dokku source.
#
# Output modes:
# - fancy: stdout is a terminal; shows a spinner and a live tail of command output for the running step
# - plain: stdout is not a terminal (CI, Vagrant, piping to a file); one line when each step starts and ends
# - verbose: DOKKU_VERBOSE=1 or TRACE=1; streams all command output
# All command output is also written to a timestamped log file under /var/log/dokku.
# Set NO_COLOR=1 to disable colors.

# We wrap this whole script in functions, so that we won't execute
# until the entire script is downloaded.
# That's good because it prevents our output overlapping with wget's.
# It also means that we can't run a partially downloaded script.

SUPPORTED_VERSIONS="Debian [11, 12, 13], Ubuntu [22.04, 24.04, 26.04]"
DOKKU_DOCS_URL="https://dokku.com/docs"

UI_MODE="plain"
UI_COLOR="false"
UI_UTF8="false"
UI_WIDTH=80
UI_TAIL_LINES=4
UI_TMP=""
UI_LOG_FILE=""
UI_NOW=0
UI_DURATION=""
UI_SANITIZED=""
UI_INSTALL_START=0
UI_STEP_INDEX=0
UI_STEP_TOTAL=0
UI_STEP_TITLE=""
UI_STEP_START=0
UI_STEP_PID=""
UI_FRAME=0
UI_FRAME_DRAWN="false"
UI_STEP_NAMES=()
UI_STEP_DURATIONS=()

log-fail() {
  declare desc="log fail formatter"
  echo "$@" 1>&2
  exit 1
}

in-array() {
  declare desc="return true if value (\$1) is in list (all other arguments)"

  local e
  for e in "${@:2}"; do
    [[ "$e" == "$1" ]] && return 0
  done
  return 1
}

ui-now() {
  declare desc="store the current time in microseconds in UI_NOW"
  UI_NOW="${EPOCHREALTIME/[.,]/}"
  if [[ -z "$UI_NOW" ]]; then
    UI_NOW="$(date +%s)000000"
  fi
}

ui-format-duration() {
  declare desc="store a human readable form of a microsecond duration in UI_DURATION"
  local micros="$1" style="$2"
  local tenths=$((micros / 100000)) seconds=$((micros / 1000000))

  if [[ "$seconds" -lt 60 ]]; then
    if [[ "$style" == "whole" ]]; then
      UI_DURATION="${seconds}s"
    else
      UI_DURATION="$((tenths / 10)).$((tenths % 10))s"
    fi
  elif [[ "$seconds" -lt 3600 ]]; then
    printf -v UI_DURATION '%dm %02ds' $((seconds / 60)) $((seconds % 60))
  else
    printf -v UI_DURATION '%dh %02dm' $((seconds / 3600)) $((seconds % 3600 / 60))
  fi
}

ui-color() {
  declare desc="print the escape sequence for a color role at a color depth"
  local role="$1" depth="$2" truecolor xterm basic

  case "$role" in
    primary) truecolor="42;143;189" xterm=31 basic=34 ;;
    teal) truecolor="7;203;207" xterm=44 basic=36 ;;
    muted) truecolor="151;166;181" xterm=109 basic=2 ;;
    success) truecolor="51;204;51" xterm=77 basic=32 ;;
    warning) truecolor="255;215;0" xterm=220 basic=33 ;;
    error) truecolor="236;99;80" xterm=203 basic=31 ;;
  esac

  case "$depth" in
    24) printf '\033[38;2;%sm' "$truecolor" ;;
    256) printf '\033[38;5;%sm' "$xterm" ;;
    *) printf '\033[%sm' "$basic" ;;
  esac
}

ui-color-depth() {
  declare desc="print the color depth supported by the terminal"
  local colors

  case "$COLORTERM" in
    truecolor | 24bit)
      echo 24
      return
      ;;
  esac

  case "$TERM" in
    *-direct* | xterm-ghostty | xterm-kitty | alacritty* | wezterm* | foot*)
      echo 24
      return
      ;;
    *256color*)
      echo 256
      return
      ;;
  esac

  colors="$(tput colors 2>/dev/null || echo 0)"
  if [[ "$colors" -ge 256 ]]; then
    echo 256
  else
    echo 16
  fi
}

ui-init-colors() {
  declare desc="set the color escape sequences for the current terminal"
  local depth

  C_RESET="" C_BOLD="" C_PRIMARY="" C_TEAL="" C_MUTED="" C_SUCCESS="" C_WARNING="" C_ERROR=""
  [[ "$UI_COLOR" == "true" ]] || return 0

  depth="$(ui-color-depth)"
  C_RESET=$'\033[0m'
  C_BOLD=$'\033[1m'
  C_PRIMARY="$(ui-color primary "$depth")"
  C_TEAL="$(ui-color teal "$depth")"
  C_MUTED="$(ui-color muted "$depth")"
  C_SUCCESS="$(ui-color success "$depth")"
  C_WARNING="$(ui-color warning "$depth")"
  C_ERROR="$(ui-color error "$depth")"
}

ui-update-width() {
  declare desc="store the terminal width in UI_WIDTH"
  local size

  size="$({ stty size </dev/tty; } 2>/dev/null || true)"
  UI_WIDTH="${size#* }"
  if [[ ! "$UI_WIDTH" =~ ^[0-9]+$ ]] || [[ "$UI_WIDTH" -lt 20 ]]; then
    UI_WIDTH="${COLUMNS:-80}"
  fi
}

ui-init() {
  declare desc="pick the output mode, colors, glyphs, and log file"
  local mode="$1" charset="${LC_ALL:-${LC_CTYPE:-$LANG}}"

  if [[ -n "$mode" ]]; then
    UI_MODE="$mode"
  elif [[ -n "$TRACE" ]] || [[ -n "$DOKKU_VERBOSE" && "$DOKKU_VERBOSE" != "0" && "$DOKKU_VERBOSE" != "false" ]]; then
    UI_MODE="verbose"
  elif [[ -t 1 ]] && [[ -n "$TERM" ]] && [[ "$TERM" != "dumb" ]]; then
    UI_MODE="fancy"
  fi

  if [[ "$UI_MODE" != "plain" ]] && [[ -t 1 ]] && [[ -z "$NO_COLOR" ]] && [[ "$TERM" != "dumb" ]]; then
    UI_COLOR="true"
  fi

  shopt -s nocasematch
  if [[ "$charset" =~ utf-?8 ]]; then
    UI_UTF8="true"
  fi
  shopt -u nocasematch

  if [[ "$UI_UTF8" == "true" ]]; then
    UI_SPINNER=("⠋" "⠙" "⠹" "⠸" "⠼" "⠴" "⠦" "⠧" "⠇" "⠏")
    UI_GLYPH_ELLIPSIS="…"
  else
    UI_SPINNER=("|" "/" "-" "\\")
    UI_GLYPH_ELLIPSIS="..."
  fi

  ui-init-colors
  ui-update-width
  UI_TMP="$(mktemp -d)"
  ui-init-log
  ui-now
  UI_INSTALL_START="$UI_NOW"

  trap ui-cleanup EXIT
  trap ui-interrupt INT TERM
  if [[ "$UI_MODE" == "fancy" ]]; then
    trap ui-update-width WINCH
  fi
}

ui-init-log() {
  declare desc="create the install log file"
  local name

  name="install-$(date +%Y%m%d-%H%M%S).log"
  if mkdir -p /var/log/dokku 2>/dev/null && : >"/var/log/dokku/$name" 2>/dev/null; then
    UI_LOG_FILE="/var/log/dokku/$name"
  else
    UI_LOG_FILE="${TMPDIR:-/tmp}/dokku-$name"
    : >"$UI_LOG_FILE"
  fi
}

ui-log() {
  declare desc="write a timestamped line to the install log"
  TZ=UTC0 printf '%(%Y-%m-%dT%H:%M:%SZ)T %s\n' -1 "$*" >>"$UI_LOG_FILE"
}

ui-cleanup() {
  declare desc="stop any running step and restore the terminal"
  ui-kill-step
  if [[ "$UI_MODE" == "fancy" ]]; then
    printf '\033[?25h'
  fi
  if [[ -n "$UI_TMP" ]]; then
    rm -rf "$UI_TMP"
  fi
}

ui-interrupt() {
  declare desc="handle ctrl-c and termination signals"
  trap - INT TERM
  ui-kill-step
  if [[ "$UI_MODE" == "fancy" ]] && [[ "$UI_FRAME_DRAWN" == "true" ]]; then
    printf '\r\033[J'
  fi
  ui-log "==> install interrupted"
  ui-error "Installation interrupted"
  ui-error "Full log: $UI_LOG_FILE"
  exit 130
}

ui-process-tree() {
  declare desc="print a process and all of its descendants"
  local pid="$1" child children=""

  echo "$pid"
  children="$(cat /proc/"$pid"/task/*/children 2>/dev/null || pgrep -P "$pid" 2>/dev/null || true)"
  for child in $children; do
    ui-process-tree "$child"
  done
}

ui-kill-step() {
  declare desc="kill the running step and everything it started"
  local pid pids=() child

  [[ -n "$UI_STEP_PID" ]] || return 0
  for child in $(cat /proc/$$/task/*/children 2>/dev/null || pgrep -P $$ 2>/dev/null || true); do
    while read -r pid; do
      pids+=("$pid")
    done < <(ui-process-tree "$child")
  done
  if [[ "${#pids[@]}" -gt 0 ]]; then
    kill "${pids[@]}" 2>/dev/null || true
  fi
  UI_STEP_PID=""
}

ui-info1() {
  declare desc="print a step line in the style of dokku_log_info1"
  printf '%s----->%s %s\n' "$C_PRIMARY" "$C_RESET" "$*"
}

ui-info2() {
  declare desc="print a heading in the style of dokku_log_info2"
  printf '%s=====>%s %s\n' "$C_PRIMARY" "$C_RESET" "$*"
}

ui-verbose() {
  declare desc="print an indented line in the style of dokku_log_verbose"
  printf '       %s\n' "$*"
}

ui-muted() {
  declare desc="print an indented, muted line"
  printf '       %s%s%s\n' "$C_MUTED" "$*" "$C_RESET"
}

ui-warn() {
  declare desc="print a warning in the style of dokku_log_warn"
  printf '%s !     %s%s\n' "$C_WARNING" "$*" "$C_RESET"
}

ui-error() {
  declare desc="print an error in the style of dokku_log_fail"
  printf '%s !     %s%s\n' "$C_ERROR" "$*" "$C_RESET"
}

ui-header() {
  declare desc="print the installer header"
  local target="$1" os="$2"

  ui-info2 "${C_BOLD}Installing $target on $os${C_RESET}"
  ui-muted "Log file: $UI_LOG_FILE"
  ui-log "==> installing $target on $os"
}

ui-truncate() {
  declare desc="truncate text to a maximum number of characters"
  local text="$1" max="$2"

  if [[ "$max" -le 0 ]]; then
    echo ""
  elif [[ "${#text}" -gt "$max" ]]; then
    echo "${text:0:max-1}${UI_GLYPH_ELLIPSIS:0:1}"
  else
    echo "$text"
  fi
}

ui-step-name() {
  declare desc="print the counter and title of the running step"
  printf '%s[%s/%s]%s %s%s%s' "$C_MUTED" "$UI_STEP_INDEX" "$UI_STEP_TOTAL" "$C_RESET" "$C_BOLD" "$UI_STEP_TITLE" "$C_RESET"
}

ui-draw-frame() {
  declare desc="draw the running step and its live output"
  local spinner lines=() line width

  ui-now
  ui-format-duration $((UI_NOW - UI_STEP_START)) whole
  spinner="${UI_SPINNER[UI_FRAME % ${#UI_SPINNER[@]}]}"
  UI_FRAME=$((UI_FRAME + 1))
  width=$((UI_WIDTH - 1))

  if [[ "$UI_TAIL_LINES" -gt 0 ]] && [[ -s "$UI_TMP/step.tail" ]]; then
    mapfile -t lines < <(tail -n "$UI_TAIL_LINES" "$UI_TMP/step.tail")
  fi

  printf '\r%s----%s%s  ' "$C_PRIMARY" "$spinner" "$C_RESET"
  ui-step-name
  printf ' %s(%s)%s\033[K' "$C_MUTED" "$UI_DURATION" "$C_RESET"
  for line in "${lines[@]}"; do
    line="$(ui-truncate "$line" $((width - 7)))"
    printf '\n       %s%s%s\033[K' "$C_MUTED" "$line" "$C_RESET"
  done
  printf '\033[J'

  if [[ "${#lines[@]}" -gt 0 ]]; then
    printf '\033[%dA' "${#lines[@]}"
  fi
  printf '\r'
  UI_FRAME_DRAWN="true"
}

ui-stream() {
  declare desc="timestamp step output into the log, the live tail, and verbose output"
  local start="$1" line display

  while IFS= read -r line || [[ -n "$line" ]]; do
    line="${line%$'\r'}"
    line="${line##*$'\r'}"
    ui-sanitize "$line"
    display="$UI_SANITIZED"
    ui-now
    ui-format-duration $((UI_NOW - start))
    TZ=UTC0 printf '%(%Y-%m-%dT%H:%M:%SZ)T %s\n' -1 "$line" >>"$UI_LOG_FILE"
    printf '%s\n' "$display" >>"$UI_TMP/step.raw"
    printf '+%s %s\n' "$UI_DURATION" "$display" >>"$UI_TMP/step.tail"
    if [[ "$UI_MODE" == "verbose" ]]; then
      printf '       %s+%s%s %s\n' "$C_MUTED" "$UI_DURATION" "$C_RESET" "$display"
    fi
  done
}

ui-sanitize() {
  declare desc="store a line without escape sequences or control characters in UI_SANITIZED"
  local line="$1" pattern=$'\033''\[[0-9;?]*[A-Za-z]'

  line="${line//$'\t'/    }"
  while [[ "$line" =~ $pattern ]]; do
    line="${line/"${BASH_REMATCH[0]}"/}"
  done
  UI_SANITIZED="${line//[[:cntrl:]]/}"
}

ui-result() {
  declare desc="record a piece of the finished step's detail text"
  echo "$*" >>"$UI_TMP/step.result"
}

ui-note() {
  declare desc="record a warning shown under the current step"
  echo "$*" >>"$UI_TMP/step.notes"
}

ui-summary-note() {
  declare desc="record a warning shown at the end of the install"
  local line
  for line in "$@"; do
    echo "$line" >>"$UI_TMP/summary.notes"
  done
  echo "" >>"$UI_TMP/summary.notes"
}

ui-run() {
  declare desc="print a command and run it"
  echo "\$ $*"
  "$@"
}

ui-action() {
  declare desc="run a command as a timed part of a step"
  local label="$1" start
  shift

  ui-now
  start="$UI_NOW"
  "$@"
  ui-now
  ui-format-duration $((UI_NOW - start))
  ui-result "$label $UI_DURATION"
}

ui-join() {
  declare desc="join lines from a file with commas"
  local file="$1" line output=""

  [[ -s "$file" ]] || return 0
  while IFS= read -r line; do
    if [[ -z "$output" ]]; then
      output="$line"
    else
      output="$output, $line"
    fi
  done <"$file"
  echo "$output"
}

ui-step() {
  declare desc="run a function as a numbered install step"
  local title="$1" detail="$2" rc=1 duration result line
  shift 2

  UI_STEP_INDEX=$((UI_STEP_INDEX + 1))
  UI_STEP_TITLE="$title"
  UI_FRAME=0
  UI_FRAME_DRAWN="false"
  : >"$UI_TMP/step.tail"
  : >"$UI_TMP/step.raw"
  : >"$UI_TMP/step.result"
  : >"$UI_TMP/step.notes"
  rm -f "$UI_TMP/step.rc"

  ui-log "==> step start [$UI_STEP_INDEX/$UI_STEP_TOTAL] $title"
  if [[ "$UI_MODE" == "fancy" ]]; then
    printf '\033[?25l'
  else
    ui-info1 "$(ui-step-name)"
  fi

  ui-now
  UI_STEP_START="$UI_NOW"
  {
    set +e
    (
      set -e
      "$@"
    ) </dev/null 2>&1
    echo "$?" >"$UI_TMP/step.rc"
  } | ui-stream "$UI_STEP_START" &
  UI_STEP_PID="$!"

  if [[ "$UI_MODE" == "fancy" ]]; then
    while kill -0 "$UI_STEP_PID" 2>/dev/null; do
      ui-draw-frame
      sleep 0.1
    done
  fi
  wait "$UI_STEP_PID" || true
  UI_STEP_PID=""
  if [[ -s "$UI_TMP/step.rc" ]]; then
    rc="$(<"$UI_TMP/step.rc")"
  fi

  ui-now
  ui-format-duration $((UI_NOW - UI_STEP_START))
  duration="$UI_DURATION"
  UI_STEP_NAMES+=("$title")
  UI_STEP_DURATIONS+=($((UI_NOW - UI_STEP_START)))
  ui-log "==> step end [$UI_STEP_INDEX/$UI_STEP_TOTAL] $title ($duration, exit $rc)"
  result="$(ui-join "$UI_TMP/step.result")"
  result="${result:-$detail}"

  if [[ "$UI_MODE" == "fancy" ]]; then
    printf '\r\033[J\033[?25h'
    if [[ "$rc" == "0" ]]; then
      ui-info1 "$(ui-step-name) $C_MUTED($duration)$C_RESET"
      [[ -z "$result" ]] || ui-muted "$result"
    fi
  elif [[ "$rc" == "0" ]]; then
    ui-muted "${result:+$result }($duration)"
  fi

  if [[ -s "$UI_TMP/step.notes" ]]; then
    while IFS= read -r line; do
      ui-warn "$line"
    done <"$UI_TMP/step.notes"
  fi

  if [[ "$rc" != "0" ]]; then
    ui-fail "$title" "$duration"
  fi
}

ui-fail() {
  declare desc="print the output of a failed step and exit"
  local title="$1" duration="$2" line

  ui-error "[$UI_STEP_INDEX/$UI_STEP_TOTAL] $title failed after $duration. Output from this step:"
  while IFS= read -r line; do
    ui-verbose "$line"
  done <"$UI_TMP/step.raw"
  ui-error "Full log: $UI_LOG_FILE"
  ui-error "Help: $DOKKU_DOCS_URL/getting-started/troubleshooting/"
  ui-log "==> install failed"
  exit 1
}

ui-countdown() {
  declare desc="show a warning and wait before continuing"
  local short="$1" seconds="$2" remaining line count=0
  shift 2

  ui-log "==> notice: $short (waiting ${seconds}s)"
  ui-warn "$1"
  for line in "${@:2}"; do
    printf '%s       %s%s\n' "$C_WARNING" "$line" "$C_RESET"
    count=$((count + 1))
  done

  if [[ "$UI_MODE" != "fancy" ]]; then
    printf '%s       Continuing in %ss...%s\n' "$C_WARNING" "$seconds" "$C_RESET"
    sleep "$seconds"
    return 0
  fi

  printf '\033[?25l'
  for ((remaining = seconds; remaining > 0; remaining--)); do
    printf '\r%s       Continuing in %ss%s%s\033[K' "$C_WARNING" "$remaining" "$UI_GLYPH_ELLIPSIS" "$C_RESET"
    sleep 1
  done
  printf '\033[%dA\r\033[J' $((count + 1))
  ui-warn "$short Waited ${seconds}s."
  printf '\033[?25h'
}

ui-slowest-steps() {
  declare desc="print the three slowest steps"
  local i j output="" picked=() slowest

  for ((j = 0; j < 3 && j < ${#UI_STEP_NAMES[@]}; j++)); do
    slowest=-1
    for i in "${!UI_STEP_NAMES[@]}"; do
      in-array "$i" "${picked[@]}" && continue
      if [[ "$slowest" -lt 0 ]] || [[ "${UI_STEP_DURATIONS[i]}" -gt "${UI_STEP_DURATIONS[slowest]}" ]]; then
        slowest="$i"
      fi
    done
    picked+=("$slowest")
    ui-format-duration "${UI_STEP_DURATIONS[slowest]}"
    [[ -z "$output" ]] || output+=", "
    output+="${UI_STEP_NAMES[slowest]} $UI_DURATION"
  done
  echo "$output"
}

ui-summary() {
  declare desc="print the install summary"
  local version="$1" domain="$2" keys="$3" host="$4" total line first="true"

  ui-now
  ui-format-duration $((UI_NOW - UI_INSTALL_START))
  total="$UI_DURATION"

  ui-info2 "${C_TEAL}${C_BOLD}Dokku $version installed in $total${C_RESET}"
  ui-verbose "$(printf '%-15s %s' "Global domain:" "${domain:-not set}")"
  ui-verbose "$(printf '%-15s %s' "SSH keys:" "${keys:-none}")"
  ui-verbose "$(printf '%-15s %s' "Slowest steps:" "$(ui-slowest-steps)")"
  ui-verbose "$(printf '%-15s %s' "Log file:" "$UI_LOG_FILE")"
  ui-info2 "Deploy an app"
  ui-verbose "git remote add dokku dokku@$host:my-app"
  ui-verbose "git push dokku main"

  if [[ -s "$UI_TMP/summary.notes" ]]; then
    while IFS= read -r line; do
      if [[ -z "$line" ]]; then
        first="true"
      elif [[ "$first" == "true" ]]; then
        ui-warn "$line"
        first="false"
      else
        ui-verbose "$line"
      fi
    done <"$UI_TMP/summary.notes"
  fi
  ui-log "==> install complete ($total)"
}

fetch-url() {
  declare desc="download a url to stdout with curl or wget"
  local url="$1"

  if command -v curl &>/dev/null; then
    curl -fsSL "$url"
  elif command -v wget &>/dev/null; then
    wget -nv -O - "$url"
  else
    log-fail "This installation script requires curl or wget"
  fi
}

package-installed() {
  declare desc="return true if a debian package is installed"
  local package="$1"
  dpkg-query -W -f='${Status}' "$package" 2>/dev/null | grep -q "ok installed"
}

docker-version() {
  declare desc="print the installed docker version"
  local version

  version="$(docker --version 2>/dev/null || true)"
  version="${version#Docker version }"
  version="${version%%,*}"
  echo "v$version"
}

dokku-os-id() {
  declare desc="print the apt codename used for the dokku package repository"
  local os_id

  os_id="$(lsb_release -cs 2>/dev/null || echo "noble")"
  if ! in-array "$DOKKU_DISTRO" "debian" "ubuntu" "raspbian"; then
    echo "noble"
  elif [[ "$DOKKU_DISTRO" == "ubuntu" ]]; then
    in-array "$os_id" "jammy" "noble" && echo "$os_id" || echo "noble"
  else
    in-array "$os_id" "bullseye" "bookworm" "trixie" && echo "$os_id" || echo "bookworm"
  fi
}

required-packages() {
  declare desc="print the packages the installer needs on this distribution"
  case "$DOKKU_DISTRO" in
    debian)
      if [[ "$DOKKU_DISTRO_VERSION" -lt "13" ]]; then
        echo "gpg-agent lsb-release software-properties-common"
      else
        echo "gpg-agent lsb-release"
      fi
      ;;
    ubuntu)
      echo "gpg-agent software-properties-common"
      ;;
  esac
}

resolve-install-method() {
  declare desc="decide whether to install from a package or from source"
  local semver major minor patch use_plugin=false

  DOKKU_INSTALL_METHOD="package"
  DOKKU_PACKAGE_VERSION=""
  DOKKU_SOURCE_CHECKOUT=""
  DOKKU_DEPENDENCIES_COMMAND="plugin:install-dependencies"

  if [[ -n $DOKKU_BRANCH ]]; then
    DOKKU_INSTALL_METHOD="source"
    DOKKU_SOURCE_CHECKOUT="origin/$DOKKU_BRANCH"
    return 0
  fi

  [[ -n $DOKKU_TAG ]] || return 0

  semver="${DOKKU_TAG//v/}"
  major=$(echo "$semver" | awk '{split($0,a,"."); print a[1]}')
  minor=$(echo "$semver" | awk '{split($0,a,"."); print a[2]}')
  patch=$(echo "$semver" | awk '{split($0,a,"."); print a[3]}')

  if [[ "$major" -eq "0" ]] && [[ "$minor" -ge "4" ]] && [[ "$patch" -ge "0" ]]; then
    use_plugin=true
  elif [[ "$major" -ge "1" ]]; then
    use_plugin=true
  fi

  if [[ "$major" -eq "0" ]] && [[ "$minor" -eq "3" ]] && [[ "$patch" -ge "13" ]]; then
    DOKKU_PACKAGE_VERSION="$semver"
    DOKKU_DEPENDENCIES_COMMAND="plugins-install-dependencies"
  elif [[ "$use_plugin" == "true" ]]; then
    DOKKU_PACKAGE_VERSION="$semver"
  else
    DOKKU_INSTALL_METHOD="source"
    DOKKU_SOURCE_CHECKOUT="$DOKKU_TAG"
  fi
}

check-environment() {
  declare desc="check the host for install requirements"
  local memory

  hostname -f &>/dev/null || {
    log-fail "This installation script requires that you have a hostname set for the instance. Please set a hostname for 127.0.0.1 in your /etc/hosts"
  }
  echo "hostname: $(hostname -f)"

  if [[ "$DOKKU_INSTALL_METHOD" == "package" ]]; then
    if ! in-array "$DOKKU_DISTRO_VERSION" "22.04" "24.04" "26.04" "10" "11" "12" "13"; then
      log-fail "Unsupported Linux distribution. Only the following versions are supported: $SUPPORTED_VERSIONS"
    fi
  fi

  memory=$(grep MemTotal /proc/meminfo | awk '{print $2}')
  echo "memory: $memory kB"
  if [[ "$memory" -lt 1003600 ]]; then
    ui-note "Less than 1024MB of memory. Builds may fail without swap."
    ui-summary-note "Less than 1024MB of memory. Add swap before building apps:" "$DOKKU_DOCS_URL/getting-started/advanced-installation/#vms-with-less-than-1-gb-of-memory"
  fi
}

enable-universe() {
  declare desc="enable the ubuntu universe repository"
  ui-run add-apt-repository -y universe
  ui-run apt-get update -qq
}

install-requirements() {
  declare desc="install the packages the installer depends on"
  local missing=() package

  for package in $(required-packages); do
    if [[ "$package" == "lsb-release" ]]; then
      command -v lsb_release &>/dev/null || missing+=("$package")
    elif ! package-installed "$package"; then
      missing+=("$package")
    fi
  done

  if [[ "${#missing[@]}" -gt 0 ]]; then
    ui-action "apt update" ui-run apt-get update -qq
    ui-action "packages" ui-run apt-get -qq -y --no-install-recommends install "${missing[@]}"
  else
    ui-result "packages already installed"
  fi

  if [[ "$DOKKU_DISTRO" == "ubuntu" ]]; then
    ui-action "universe" enable-universe
  fi
}

install-docker() {
  declare desc="install docker via get.docker.com"
  if command -v docker &>/dev/null; then
    echo "docker already installed: $(docker --version)"
    ui-result "already installed $(docker-version)"
    return 0
  fi

  export CHANNEL=stable
  echo "\$ fetch https://get.docker.com/ | sh"
  fetch-url https://get.docker.com/ | sh
  ui-result "$(docker-version)"
  ui-result "get.docker.com"
}

add-dokku-repository() {
  declare desc="add the dokku apt repository"
  local os_id

  case "$DOKKU_DISTRO" in
    debian | ubuntu) ;;
    *)
      log-fail "Unsupported Linux distribution. For manual installation instructions, consult $DOKKU_DOCS_URL/getting-started/advanced-installation/"
      ;;
  esac

  os_id="$(dokku-os-id)"
  echo "\$ fetch https://packagecloud.io/dokku/dokku/gpgkey > /etc/apt/trusted.gpg.d/dokku.asc"
  fetch-url https://packagecloud.io/dokku/dokku/gpgkey | sudo tee /etc/apt/trusted.gpg.d/dokku.asc
  echo "deb https://packagecloud.io/dokku/dokku/$DOKKU_DISTRO/ $os_id main" | tee /etc/apt/sources.list.d/dokku.list
  ui-run apt-get update -qq
}

install-dokku-package() {
  declare desc="install the dokku debian package"
  local version="$1"
  local NO_INSTALL_RECOMMENDS=${DOKKU_NO_INSTALL_RECOMMENDS:=""}

  if [[ -n $DOKKU_DOCKERFILE ]]; then
    NO_INSTALL_RECOMMENDS=" --no-install-recommends "
  fi

  if [[ -n $DOKKU_VHOST_ENABLE ]]; then
    echo "dokku dokku/vhost_enable boolean $DOKKU_VHOST_ENABLE" | sudo debconf-set-selections
  fi
  if [[ -n $DOKKU_HOSTNAME ]]; then
    echo "dokku dokku/hostname string $DOKKU_HOSTNAME" | sudo debconf-set-selections
  fi
  if [[ -n $DOKKU_SKIP_KEY_FILE ]]; then
    echo "dokku dokku/skip_key_file boolean $DOKKU_SKIP_KEY_FILE" | sudo debconf-set-selections
  fi
  if [[ -n $DOKKU_KEY_FILE ]]; then
    echo "dokku dokku/key_file string $DOKKU_KEY_FILE" | sudo debconf-set-selections
  fi
  if [[ -n $DOKKU_NGINX_ENABLE ]]; then
    echo "dokku dokku/nginx_enable string $DOKKU_NGINX_ENABLE" | sudo debconf-set-selections
  fi

  if [[ -n $version ]]; then
    ui-run apt-get -qq -y $NO_INSTALL_RECOMMENDS install "dokku=$version"
  else
    ui-run apt-get -qq -y $NO_INSTALL_RECOMMENDS install dokku
  fi
}

install-dokku-dependencies() {
  declare desc="install the dependencies of the core dokku plugins"
  if [[ "$DOKKU_DEPENDENCIES_COMMAND" == "plugins-install-dependencies" ]]; then
    ui-run dokku plugins-install-dependencies
  else
    ui-run sudo dokku plugin:install-dependencies --core
  fi
}

install-source-requirements() {
  declare desc="install the packages needed to build dokku from source"
  if ! command -v apt-get &>/dev/null; then
    log-fail "This installation script requires apt-get. For manual installation instructions, consult $DOKKU_DOCS_URL/getting-started/advanced-installation/"
  fi

  ui-run apt-get -qq -y --no-install-recommends install sudo git make
  case "$DOKKU_DISTRO" in
    debian)
      if [[ "$DOKKU_DISTRO_VERSION" -lt "13" ]] && ! package-installed software-properties-common; then
        ui-run apt-get -qq -y --no-install-recommends install software-properties-common
      fi
      ;;
    ubuntu)
      if ! package-installed software-properties-common; then
        ui-run apt-get -qq -y --no-install-recommends install software-properties-common
      fi
      ;;
  esac
}

clone-dokku-source() {
  declare desc="clone the dokku repository and check out the requested ref"
  local checkout="$1"

  cd /root
  if [[ ! -d /root/dokku ]]; then
    ui-run git clone "$DOKKU_REPO" /root/dokku
  fi

  cd /root/dokku
  ui-run git fetch origin
  if [[ -n $checkout ]]; then
    ui-run git checkout "$checkout"
  fi
  ui-result "$(git rev-parse --short HEAD)"
}

build-dokku-source() {
  declare desc="build and install dokku from source"
  cd /root/dokku
  ui-run make install
}

primary-ip() {
  declare desc="print the primary ipv4 address of the server"
  local route

  route="$(ip -4 route get 1.1.1.1 2>/dev/null || true)"
  if [[ "$route" =~ src\ ([0-9.]+) ]]; then
    echo "${BASH_REMATCH[1]}"
    return 0
  fi
  route="$(hostname -I 2>/dev/null || true)"
  echo "${route%% *}"
}

add-ssh-keys() {
  declare desc="add the keys from a file as admin keys"
  local key_file="$1" line key name count=0

  while IFS= read -r line || [[ -n "$line" ]]; do
    if [[ "$line" =~ ((ssh-(rsa|ed25519|dss)|ecdsa-sha2-[a-z0-9]+|sk-[a-z0-9@.-]+)\ [A-Za-z0-9+/=]+) ]]; then
      key="${BASH_REMATCH[1]}"
      count=$((count + 1))
      name="admin"
      [[ "$count" -eq 1 ]] || name="admin-$count"
      echo "\$ dokku ssh-keys:add $name"
      if ! echo "$key" | dokku ssh-keys:add "$name"; then
        echo "skipping a key from $key_file that could not be added"
      fi
    fi
  done <"$key_file"
}

set-global-domain() {
  declare desc="set the global domain to the server's ip"
  local ip="$1"
  ui-run dokku domains:set-global "$ip"
}

configure-dokku() {
  declare desc="add ssh keys and a global domain when they are missing"
  local dokku_root="${DOKKU_ROOT:-/home/dokku}" key_file="/root/.ssh/authorized_keys" ip

  if [[ -s "$dokku_root/.ssh/authorized_keys" ]]; then
    echo "ssh keys already present, skipping"
    ui-result "ssh keys already set"
  elif [[ "$DOKKU_SKIP_KEY_FILE" == "true" ]]; then
    echo "DOKKU_SKIP_KEY_FILE is set, skipping ssh keys"
    ui-result "admin key skipped"
  elif [[ ! -s "$key_file" ]]; then
    echo "no keys found in $key_file, skipping"
    ui-result "admin key skipped"
  else
    ui-action "admin key" add-ssh-keys "$key_file"
    echo "$key_file" >"$UI_TMP/ssh-key-source"
  fi

  if [[ -s "$dokku_root/VHOST" ]]; then
    echo "global domain already set to $(<"$dokku_root/VHOST")"
    ui-result "global domain already set"
  else
    ip="$(primary-ip)"
    if [[ -n "$ip" ]]; then
      ui-action "global domain" set-global-domain "$ip"
    else
      ui-result "global domain skipped"
    fi
  fi
}

ssh-key-names() {
  declare desc="print the names of the keys allowed to push to dokku"
  local dokku_root="${DOKKU_ROOT:-/home/dokku}" line names=()

  [[ -f "$dokku_root/.ssh/authorized_keys" ]] || return 0
  while IFS= read -r line; do
    if [[ "$line" =~ NAME=\\\"([^\\]+)\\\" ]]; then
      names+=("${BASH_REMATCH[1]}")
    fi
  done <"$dokku_root/.ssh/authorized_keys"
  echo "${names[*]}"
}

print-summary() {
  declare desc="print the ready summary with what was configured"
  local dokku_root="${DOKKU_ROOT:-/home/dokku}" version domain="" keys host

  version="$(dokku version 2>/dev/null || true)"
  version="v${version##* }"
  if [[ -s "$dokku_root/VHOST" ]]; then
    domain="$(<"$dokku_root/VHOST")"
  fi

  keys="$(ssh-key-names)"
  if [[ -n "$keys" ]] && [[ -s "$UI_TMP/ssh-key-source" ]]; then
    keys="$keys (from $(<"$UI_TMP/ssh-key-source"))"
  elif [[ -z "$keys" ]]; then
    ui-summary-note "No SSH key was added. Add the key you will deploy with:" "echo 'CONTENTS_OF_ID_RSA_PUB_FILE' | dokku ssh-keys:add admin"
  fi

  if [[ -z "$domain" ]]; then
    ui-summary-note "No global domain is set. Set one to route apps by name:" "dokku domains:set-global dokku.me"
  fi

  host="${domain:-$(primary-ip)}"
  ui-summary "$version" "$domain" "$keys" "${host:-SERVER_IP}"
}

nginx-notice() {
  declare desc="warn that a first time install removes nginx sites"
  if command -v dokku &>/dev/null; then
    return 0
  fi
  ui-countdown "First install removes nginx sites-enabled files." 10 \
    "Installing Dokku for the first time removes files in nginx's sites-enabled" \
    "directory. Restore any removed files after the installation is complete."
}

linode-notice() {
  declare desc="warn linode users about aufs messages from docker"
  if command -v docker &>/dev/null || ! uname -r | grep -q linode; then
    return 0
  fi
  ui-countdown "Docker may warn about missing AUFS support on Linode." 10 \
    "Using Linode? Docker may complain about missing AUFS support." \
    "You can safely ignore this warning."
}

install-target() {
  declare desc="print what is being installed for the header"
  if [[ -n $DOKKU_BRANCH ]]; then
    echo "Dokku from branch $DOKKU_BRANCH"
  elif [[ -n $DOKKU_TAG ]]; then
    echo "Dokku $DOKKU_TAG"
  else
    echo "the latest Dokku"
  fi
}

os-description() {
  declare desc="print the distribution and architecture"
  local name arch

  name="$(. /etc/os-release && echo "$NAME $VERSION_ID")"
  arch="$(dpkg --print-architecture 2>/dev/null || uname -m)"
  echo "$name ($arch)"
}

environment-detail() {
  declare desc="print the detail text for the environment step"
  local memory arch

  memory=$(grep MemTotal /proc/meminfo | awk '{print $2}')
  arch="$(dpkg --print-architecture 2>/dev/null || uname -m)"
  if [[ "$memory" -ge 1048576 ]]; then
    memory="$(((memory + 524288) / 1048576))GB"
  else
    memory="$((memory / 1024))MB"
  fi
  echo "$DOKKU_DISTRO $DOKKU_DISTRO_VERSION, $arch, $memory"
}

main() {
  declare desc="install dokku"
  local docker_detail="" version_detail

  export DOKKU_DISTRO DOKKU_DISTRO_VERSION
  DOKKU_DISTRO=$(. /etc/os-release && echo "$ID")
  DOKKU_DISTRO_VERSION=$(. /etc/os-release && echo "$VERSION_ID")

  export DEBIAN_FRONTEND=noninteractive
  export DOKKU_REPO=${DOKKU_REPO:-"https://github.com/dokku/dokku.git"}

  resolve-install-method
  ui-init
  ui-header "$(install-target)" "$(os-description)"

  if [[ "$DOKKU_INSTALL_METHOD" == "package" ]]; then
    UI_STEP_TOTAL=7
  else
    UI_STEP_TOTAL=6
  fi

  ui-step "Checking environment" "$(environment-detail)" check-environment
  ui-step "Installing required packages" "$(required-packages)" install-requirements

  if [[ "$DOKKU_INSTALL_METHOD" == "package" ]]; then
    command -v docker &>/dev/null || docker_detail="get.docker.com"
    linode-notice
    ui-step "Installing Docker Engine" "$docker_detail" install-docker
    ui-step "Adding Dokku apt repository" "packagecloud, $(dokku-os-id)" add-dokku-repository
    nginx-notice
    version_detail="${DOKKU_PACKAGE_VERSION:+v$DOKKU_PACKAGE_VERSION}"
    ui-step "Installing Dokku" "${version_detail:-latest}" install-dokku-package "$DOKKU_PACKAGE_VERSION"
    ui-step "Installing plugin dependencies" "core plugins" install-dokku-dependencies
  else
    ui-step "Installing build tools" "sudo git make" install-source-requirements
    ui-step "Cloning Dokku" "${DOKKU_SOURCE_CHECKOUT:-$DOKKU_REPO}" clone-dokku-source "$DOKKU_SOURCE_CHECKOUT"
    nginx-notice
    ui-step "Building Dokku" "make install" build-dokku-source
  fi

  ui-step "Configuring Dokku" "" configure-dokku
  print-summary
}

{
  main "$@"
  exit $?
}
