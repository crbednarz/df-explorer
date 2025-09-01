DF_AT_PROMPT=1
DF_WORKDIR="/"
DF_OPERATION=""
DF_COMMAND=""

PROMPT_COMMAND="DFPostCommand"
function DFPreCommand() {
  if [ -z "$DF_AT_PROMPT" ]; then
    return
  fi
  unset DF_AT_PROMPT
  DF_CAPTURED_COMMAND="$BASH_COMMAND"

  if [[ "$DF_CAPTURED_COMMAND" == "DFPostCommand" ]]; then
    return
  fi

  local operation="$DF_OPERATION"
  local command="$DF_COMMAND"
  if [ -z "$DF_OPERATION" ]; then
    operation=""
    command="$DF_CAPTURED_COMMAND"
  fi

  command=${command//\\/\\\\}
  command=${command//\"/\\\"}

  printf '{"operation": "%s", "command":"%s","status":"running"}\n' "$operation" "$command" >>/tmp/df-explorer/history.log
}
trap "DFPreCommand" DEBUG

function DFPostCommand() {
  local rc=$?
  DF_AT_PROMPT=1

  if [[ "$DF_CAPTURED_COMMAND" == "DFPostCommand" ]]; then
    return
  fi

  local operation="$DF_OPERATION"
  local command="$DF_COMMAND"
  if [ -z "$DF_OPERATION" ]; then
    operation=""
    command="$DF_CAPTURED_COMMAND"
  fi

  command=${command//\\/\\\\}
  command=${command//\"/\\\"}

  printf '{"operation": "%s", "command":"%s","status":"complete","rc":%d}\n' "$operation" "$command" $rc >>/tmp/df-explorer/history.log
  unset DF_CAPTURED_COMMAND
  unset DF_OPERATION
  unset DF_COMMAND
}

function RUN() {
  DF_OPERATION="RUN"
  DF_COMMAND="$*"
  sh -c "cd $DF_WORKDIR; $*"
}

function ENV() {
  DF_OPERATION="ENV"
  DF_COMMAND="$*"
  eval export $@
}

function WORKDIR() {
  DF_OPERATION="WORKDIR"
  DF_COMMAND="$*"
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
