DF_AT_PROMPT=1
DF_RUNNING_COMMAND=""
function DFPreCommand() {
  if [ -z "$DF_AT_PROMPT" ]; then
    return
  fi
  unset DF_AT_PROMPT

  DF_RUNNING_COMMAND="$BASH_COMMAND"
  printf "{\"command\":\"%q\",\"status\":\"running\"}" "$DF_RUNNING_COMMAND" >>/tmp/df-explorer/history.log
}
trap "DFPreCommand" DEBUG

function DFPostCommand() {
  local return_code=$?
  DF_AT_PROMPT=1

  printf "{\"command\":\"%q\",\"status\":\"complete\",\"rc\":$return_code}" "$DF_RUNNING_COMMAND" >>/tmp/df-explorer/history.log
  unset DF_RUNNING_COMMAND
}
PROMPT_COMMAND="DFPostCommand"

alias RUN="DF_RUN=1"
alias ENV="DF_ENV=1 export"
