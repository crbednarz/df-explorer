AT_PROMPT=1
RUNNING_BASH_COMMAND=""
function DFPreCommand() {
  if [ -z "$AT_PROMPT" ]; then
    return
  fi
  unset AT_PROMPT

  RUNNING_BASH_COMMAND="$BASH_COMMAND"
  echo "{\"command\":\"$RUNNING_BASH_COMMAND\",\"status\":\"running\"}" >>/tmp/df-explorer/history.log
}
trap "DFPreCommand" DEBUG

function DFPostCommand() {
  AT_PROMPT=1

  echo "{\"command\":\"$RUNNING_BASH_COMMAND\",\"status\":\"complete\"}" >>/tmp/df-explorer/history.log
  unset RUNNING_BASH_COMMAND
}
PROMPT_COMMAND="DFPostCommand"
