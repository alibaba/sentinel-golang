#!/bin/bash

node_crash=${1:-false}
node_count=${2:-9} # node_count can only be 1 to 9

start_process() {
  for ((i = 1; i <= node_count; i++)); do
    port=900$i
    echo "Starting: go run . --server_address=:$port"
    go run . --server_address=:$port &
    pids[$i]=$!
  done
}

stop_process() {
  for ((i = 1; i <= node_count; i++)); do
    sleep 5
    kill ${pids[$i]}
    port=900$i
    pgrep -f "hello_micro --server_address=:$port" | xargs kill
    echo "Killed process ${pids[$i]}"
  done
}

restart_process() {
  for ((i = 1; i <= node_count; i++)); do
    sleep 5
    port=900$i
    echo "Restarting: go run . --server_address=:$port"
    go run . --server_address=:$port &
  done
}

sed -E -i 's/nodeCrash = (true|false)/nodeCrash = '$node_crash'/' handler.go
if [ "$node_crash" = "true" ]; then
  start_process
  sleep 5
  stop_process
  restart_process
else
  start_process
fi

wait
