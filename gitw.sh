#!/bin/bash
DIR="$( cd "$( dirname "$(readlink -f "${BASH_SOURCE[0]}")" )" && pwd )"
# shellcheck source=/dev/null
source "${DIR}/echos"

#echo "$-"
#echo "${gitshellenv}"
#if [[ ${gitshellenv} == *i* ]]; then info "interactive";else warning "non interactive";fi

# Make sure .bashrc has alias git='gitshellenv=$- ${DIR}/gitw.sh'
# That will set gitshellenv, used here in this git wrapper script

b="${HOME}/.bashrc"
if [[ -e "${HOME}/.env" ]]; then
  b="${HOME}/.env"
fi

#echo "1='${1}'"
if [[ "${1}" == "gitwset" ]]; then
  if grep -E "^\s*?#\s*?alias git='gitshellenv" "${b}" >/dev/null; then
    sed -i "s/\s*\?#\s*\?alias git=/alias git=/g" "${b}"
    ok "gitw wrapper activated again"
    exit 0
  fi

  if ! grep -E "^\s*?alias git='gitshellenv" "${b}" >/dev/null; then
    echo "alias git='gitshellenv=\$- \"${DIR}/gitw.sh\"'" >> "${b}"
    ok "gitw wrapper set"
  else
    info "gitw wrapper /already/ set"
  fi
  exit 0
fi

if [[ "${1}" == "gitwunset" ]]; then
  if grep -E "^\s*?#\s*?alias git='gitshellenv" "${b}" >/dev/null; then
    info "gitw wrapper /already/ unset."
    exit 0
  fi
  if ! grep -E "^\s*?alias git='gitshellenv" "${b}" >/dev/null; then
    info "gitw wrapper was never set in the first place."
    exit 0
  fi
  if ! sed -i "s/\s*\?alias git=/# alias git=/g" "${b}"; then
    error "Issue commenting alias git in '${b}'"
    exit 1
  fi
  ok "gitw wrapper unset"
  exit 0
fi

if [[ "${1}" == "gitwupgrade" ]]; then
  if ! scp SANBOX:gitw/gitw "${DIR}/gitw"; then
    info "Nothing to upgrade to from SANBOX:gitw/gitw"
    exit 0
  fi
  ok "gitw wrapper upgraded to version\ngitw $("${DIR}/gitw" version)"
  exit 0
fi

if [[ "${1}" == "gitwversion" ]]; then
  ok "Displaying gitw wrapper version"
  "${DIR}/gitw" version
  exit 0
fi

# shellcheck disable=SC2154
if [[ ${gitshellenv} != *i* ]]; then
  /usr/bin/git "$@"
  exit $?
fi
#info "interactive Git: checking user"

if [[ "${GIT_AUTHOR_NAME}" != "" ]]; then
  ok "Valid Git user '${GIT_AUTHOR_NAME}/${GIT_AUTHOR_EMAIL}'"
fi
#warning "GIT_AUTHOR_NAME not set: check user identity from SSH connection"

g="${DIR}/gitw"
if ! "${g}" check; then
  "${g}"
fi
unue=$("${g}")
# info "unue='${unue}'"
un=${unue%/*}
ue=${unue#*/}
ok "Git User '${un}'/'${ue}'"

export GIT_AUTHOR_NAME="${un}"
export GIT_AUTHOR_EMAIL="${ue}"
export GIT_COMMITTER_NAME="${un}"
export GIT_COMMITTER_EMAIL="${ue}"
HOME=${DIR}/.. command git "$@"
