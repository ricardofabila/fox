#!/usr/bin/env bash

# We don't need return codes for "$(command)", only stdout is needed.
# Allow `[[ -n "$(command)" ]]`, `func "$(command)"`, pipes, etc.
# shellcheck disable=SC2312

set -u

abort() {
  printf "%s\n" "$@" >&2
  exit 1
}

# Search for the given executable in PATH (avoids a dependency on the `which` command)
which() {
  # Alias to Bash built-in command `type -P`
  type -P "$@"
}

# Fail fast with a concise message when not using bash
# Single brackets are needed here for POSIX compatibility
# shellcheck disable=SC2292
if [ -z "${BASH_VERSION:-}" ]; then
  abort "Bash is required to interpret this script."
fi

# string formatters
if [[ -t 1 ]]; then
  tty_escape() { printf "\033[%sm" "$1"; }
else
  tty_escape() { :; }
fi
tty_mkbold() { tty_escape "1;$1"; }
tty_underline="$(tty_escape "4;39")"
tty_blue="$(tty_mkbold 34)"
tty_red="$(tty_mkbold 31)"
tty_bold="$(tty_mkbold 39)"
tty_reset="$(tty_escape 0)"

shell_join() {
  local arg
  printf "%s" "$1"
  shift
  for arg in "$@"; do
    printf " "
    printf "%s" "${arg// /\ }"
  done
}

chomp() {
  printf "%s" "${1/"$'\n'"/}"
}

ohai() {
  printf "${tty_blue}==>${tty_bold} %s${tty_reset}\n" "$(shell_join "$@")"
}

warn() {
  printf "${tty_red}Warning${tty_reset}: %s\n" "$(chomp "$1")"
}

# USER isn't always set so provide a fall back for the installer and subprocesses.
if [[ -z "${USER-}" ]]; then
  USER="$(chomp "$(id -un)")"
  export USER
fi

CHMOD=("$(which "chmod")")
MKDIR=("$(which "mkdir")" "-p")
SUDO=("$(which "sudo")")
UNAME="$(which "uname")"

# First check OS.
OS="$($UNAME)"
if [[ "${OS}" == "Linux" ]]; then
  FOX_ON_LINUX=1
elif [[ "${OS}" != "Darwin" ]]; then
  abort "fox is only supported on macOS and Linux."
fi

# Required installation paths. To install elsewhere (which is unsupported)
# you can get a release from https://github.com/ricardofabila/fox and place anywhere you like.
FOX_PREFIX="/usr/local/Fox"
UNAME_MACHINE="$($UNAME -m)"

unset HAVE_SUDO_ACCESS # unset this from the environment
have_sudo_access() {
  if [[ ! -x "$(which "sudo")" ]]; then
    return 1
  fi

  if [[ -n "${SUDO_ASKPASS-}" ]]; then
    SUDO+=("-A")
  elif [[ -n "${NONINTERACTIVE-}" ]]; then
    SUDO+=("-n")
  fi

  if [[ -z "${HAVE_SUDO_ACCESS-}" ]]; then
    if [[ -n "${NONINTERACTIVE-}" ]]; then
      "${SUDO[@]}" -l mkdir &>/dev/null
    else
      "${SUDO[@]}" -v && "${SUDO[@]}" -l mkdir &>/dev/null
    fi
    HAVE_SUDO_ACCESS="$?"
  fi

  if [[ -z "${FOX_ON_LINUX-}" ]] && [[ "${HAVE_SUDO_ACCESS}" -ne 0 ]]; then
    abort "Need sudo access on macOS (e.g. the user ${USER} needs to be an Administrator)!"
  fi

  return "${HAVE_SUDO_ACCESS}"
}

execute() {
  if ! "$@"; then
    abort "$(printf "Failed during: %s" "$(shell_join "$@")")"
  fi
}

execute_sudo() {
  local -a args=("$@")
  if have_sudo_access; then
    if [[ -n "${SUDO_ASKPASS-}" ]]; then
      args=("-A" "${args[@]}")
    fi
    ohai "${SUDO[@]}" "-E" "${args[@]}"
    execute "${SUDO[@]}" "-E" "${args[@]}"
  else
    ohai "${args[@]}"
    execute "${args[@]}"
  fi
}

ring_bell() {
  # Use the shell's audible bell.
  if [[ -t 1 ]]; then
    printf "\a"
  fi
}

wait_for_user() {
  local c
  echo
  echo "Press ${tty_bold}RETURN${tty_reset}/${tty_bold}ENTER${tty_reset} to continue or any other key to abort:"
  getc c
  # we test for \r and \n because some stuff does \r instead
  if ! [[ "${c}" == $'\r' || "${c}" == $'\n' ]]; then
    exit 1
  fi
}

# Search PATH for the specified program that satisfies fox requirements
# function which is set above
# shellcheck disable=SC2230
find_tool() {
  if [[ $# -ne 1 ]]; then
    return 1
  fi

  local executable
  while read -r executable; do
    if "test_$1" "${executable}"; then
      echo "${executable}"
      break
    fi
  done < <(which -a "$1")
}

# Invalidate sudo timestamp before exiting (if it wasn't active before).
if [[ -x /usr/bin/sudo ]] && ! /usr/bin/sudo -n -v 2>/dev/null; then
  trap '/usr/bin/sudo -k' EXIT
fi

# Things can fail later if `pwd` doesn't exist.
# Also sudo prints a warning message for no good reason
cd "/usr" || exit 1

# shellcheck disable=SC2016
ohai 'Checking for `sudo` access (which may request your password)...'

if [[ -z "${FOX_ON_LINUX-}" ]]; then
  # On macOS, support 64-bit Intel and ARM
  if [[ "${UNAME_MACHINE}" != "arm64" ]] && [[ "${UNAME_MACHINE}" != "x86_64" ]]; then
    abort "fox is only supported on Intel and ARM processors!"
  fi
else
  # On Linux, support only 64-bit Intel and arm64 and ARM
  if [[ "${UNAME_MACHINE}" != "arm64" ]] && [[ "${UNAME_MACHINE}" != "x86_64" ]] && [[ "${UNAME_MACHINE}" != "aarch64" ]]; then
    abort "fox is only supported on Linux on Intel and ARM processors!"
  fi
fi

ohai "This script will install fox in:"
echo "${FOX_PREFIX}"
echo ""

# abundant thanks to -p but I like being explicit
execute_sudo "${MKDIR[@]}" "${FOX_PREFIX}"
execute_sudo "${MKDIR[@]}" "${FOX_PREFIX}/bin"
execute_sudo "${MKDIR[@]}" "${FOX_PREFIX}/temp_fox_folder"
execute_sudo "${CHMOD[@]}" "-R" "777" "${FOX_PREFIX}"

EXECUTABLE_NAME=""
if [[ -z "${FOX_ON_LINUX-}" ]]; then
  # On macOS, support 64-bit Intel and ARM
  if [[ "${UNAME_MACHINE}" == "x86_64" ]]; then
    EXECUTABLE_NAME="fox_darwin_amd64_v1"
  fi

  if [[ "${UNAME_MACHINE}" == "arm64" ]]; then
    EXECUTABLE_NAME="fox_darwin_arm64"
  fi
else
  if [[ "${UNAME_MACHINE}" == "x86_64" ]]; then
    EXECUTABLE_NAME="fox_linux_386"
  fi

  if [[ "${UNAME_MACHINE}" == "arm64" ]]; then
    EXECUTABLE_NAME="fox_linux_arm64"
  fi

  if [[ "${UNAME_MACHINE}" == "aarch64" ]]; then
    EXECUTABLE_NAME="fox_linux_amd64_v1"
  fi
fi

TAG_NAME=$(curl -s https://api.github.com/repos/ricardofabila/fox/releases/latest |
  grep "tag_name" |
  awk '{print substr($2, 2, length($2)-3)}')

execute_sudo "rm" "-f" "${FOX_PREFIX}/temp_fox_folder/${EXECUTABLE_NAME}"
execute_sudo "curl" "-L" "--output" "${FOX_PREFIX}/temp_fox_folder/${EXECUTABLE_NAME}" "https://github.com/ricardofabila/fox/releases/download/${TAG_NAME}/${EXECUTABLE_NAME}"
execute_sudo "mv" "-f" "${FOX_PREFIX}/temp_fox_folder/${EXECUTABLE_NAME}" "${FOX_PREFIX}/bin"
execute_sudo "mv" "-f" "${FOX_PREFIX}/bin/${EXECUTABLE_NAME}" "${FOX_PREFIX}/bin/fox"
execute_sudo "${CHMOD[@]}" "+rwx" "${FOX_PREFIX}/bin/fox"

echo ""

if [[ ":${PATH}:" != *":${FOX_PREFIX}/bin:"* ]]; then
  warn "${FOX_PREFIX}/bin is not in your PATH.
  Instructions on how to configure your shell for fox
  can be found in the 'Next steps' section below."
  echo ""

  ohai "Next steps:"
  USERS_SHELL=""
  if [[ "$(which "perl")" != "" ]]; then
    # shellcheck disable=SC2154
    USERS_SHELL=$(perl -e '@x=getpwuid($<); print $x[8]')
  else
    if [[ "$(which "finger")" != "" ]] && [[ "$(which "grep")" != "" ]] && [[ "$(which "cut")" != "" ]]; then
      # shellcheck disable=SC2154
      USERS_SHELL=$(finger "$USER" | grep 'Shell:*' | cut -f3 -d ":")
    fi
  fi

  shell_profile=""
  case "${USERS_SHELL}" in
  */bash*)
    if [[ -r "${HOME}/.bash_profile" ]]; then
      shell_profile="${HOME}/.bash_profile"
    else
      shell_profile="${HOME}/.profile"
    fi
    ;;
  */zsh*)
    shell_profile="${HOME}/.zprofile"
    ;;
  */fish*)
    shell_profile="${HOME}/.config/fish/config.fish"
    ;;
  esac

  if [[ "$shell_profile" == "" ]]; then
    cat <<EOS
    Couldn't find your preferred shell for your user.
    Add /usr/local/Fox/bin to your PATH.

    To know what shell you are using run:
    ${tty_bold}echo \$SHELL${tty_reset}

    Some useful links:
    ${tty_underline}https://www.baeldung.com/linux/path-variable${tty_reset}
    ${tty_underline}https://www.cyberciti.biz/faq/unix-linux-adding-path/${tty_reset}

    - In the case you are using ${tty_bold}fish${tty_reset} (as you should)
      you need this to add this to your ${tty_bold}~/.config/fish/config.fish${tty_reset}:

        ${tty_bold}fish_add_path -g /usr/local/Fox/bin${tty_reset}

      to add ${tty_blue}fox${tty_reset} to your ${tty_bold}PATH${tty_reset}

      Then you can restart your terminal.
EOS
    echo ""
    ohai "Installation successful!"
    echo ""
    exit 1
  fi

  # `which` is a shell function defined above.
  # shellcheck disable=SC2230
  if [[ "$(which fox)" != "${FOX_PREFIX}/bin/fox" ]]; then

    if [[ "$shell_profile" != "${HOME}/.config/fish/config.fish" ]]; then

      cat <<EOS
- Run these two commands in your terminal to add ${tty_blue}fox${tty_reset} to your ${tty_bold}PATH${tty_reset}.
  You can do this by adding this to your ${shell_profile}
    ${tty_bold}export PATH="/usr/local/Fox/bin:\$PATH"${tty_reset}

  Then you can restart your terminal or run:
    ${tty_bold}source ${shell_profile}${tty_reset}
EOS

    else

      cat <<EOS
- In the case you are using ${tty_bold}fish${tty_reset} (as you should)
  you need this to add this to your ${tty_bold}~/.config/fish/config.fish${tty_reset}:

    ${tty_bold}fish_add_path -g /usr/local/Fox/bin${tty_reset}

  to add ${tty_blue}fox${tty_reset} to your ${tty_bold}PATH${tty_reset}

  Then you can restart your terminal or run:
      ${tty_bold}source ${shell_profile}${tty_reset}
EOS

    fi

  fi
fi

echo ""
ohai "Installation successful!"
echo ""

ring_bell
