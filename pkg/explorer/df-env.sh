DF_AT_PROMPT=1
DF_RUNNING_COMMAND=""
DF_WORKDIR="/"
PROMPT_COMMAND="DFPostCommand"
function DFPreCommand() {
  if [ -z "$DF_AT_PROMPT" ]; then
    return
  fi
  unset DF_AT_PROMPT
  DF_RUNNING_COMMAND="$BASH_COMMAND"

  if [[ "$DF_RUNNING_COMMAND" != "DFPostCommand" ]]; then
    printf '{"command":"%q","status":"running"}\n' "$DF_RUNNING_COMMAND" >>/tmp/df-explorer/history.log
  fi
}
trap "DFPreCommand" DEBUG

function DFPostCommand() {
  local rc=$?
  DF_AT_PROMPT=1

  if [[ "$DF_RUNNING_COMMAND" != "DFPostCommand" ]]; then
    printf '{"command":"%q","status":"complete","rc":%d}\n' "$DF_RUNNING_COMMAND" $rc >>/tmp/df-explorer/history.log
  fi
  unset DF_RUNNING_COMMAND
}

function RUN() {
  sh -c "cd $DF_WORKDIR; $*"
}

function ENV() {
  eval export $@
}

function WORKDIR() {
  DF_WORKDIR="$*"
}

_df_run_complete() {
  # Current word being completed
  local cur="${COMP_WORDS[COMP_CWORD]}"
  # Skip the 'RUN' prefix to get the actual command
  local cmd="${COMP_WORDS[1]}"

  # Shift COMP_WORDS to fake completion for the actual command
  COMP_WORDS=("${COMP_WORDS[@]:1}")
  ((COMP_CWORD--))

  # Try to delegate to the real command's completion
  if complete -p "$cmd" &>/dev/null; then
    # Extract the completion function
    local compfn
    compfn=$(complete -p "$cmd" | awk '{print $3}')
    if declare -F "$compfn" &>/dev/null; then
      "$compfn"
    fi
  else
    # Fallback: file completion
    COMPREPLY=($(compgen -f -- "$cur"))
  fi
}

complete -F _df_run_complete RUN
