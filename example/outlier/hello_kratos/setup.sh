#!/bin/bash

node_crash=${1:-false}
node_count=${2:-9} # node_count can only be 1 to 9

start_process() {
  for ((i = 1; i <= node_count; i++)); do
    http_port=800$i
    grpc_port=900$i
    echo "Starting: go run . --grpc_server_address=:$grpc_port --http_server_address=:$http_port --node_crash=$node_crash"
    go run . --grpc_server_address=:$grpc_port --http_server_address=:$http_port --node_crash=$node_crash &
    pids[$i]=$!
  done
}

stop_process() {
  for ((i = 1; i <= node_count; i++)); do
    sleep 5
    kill ${pids[$i]}
    port=900$i
    pgrep -f "hello_kratos --grpc_server_address=:$port" | xargs kill
    echo "Killed process ${pids[$i]}"
  done
}

restart_process() {
  for ((i = 1; i <= node_count; i++)); do
    sleep 5
    http_port=800$i
    grpc_port=900$i
    echo "Restarting: go run . --grpc_server_address=:$grpc_port --http_server_address=:$http_port --node_crash=$node_crash"
    go run . --grpc_server_address=:$grpc_port --http_server_address=:$http_port --node_crash=$node_crash &
  done
}

if [ "$node_crash" = "true" ]; then
  start_process
  sleep 5
  stop_process
  restart_process
else
  start_process
fi

wait
