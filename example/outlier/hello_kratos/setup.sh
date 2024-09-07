#!/bin/bash
network_error=$1
for i in {1..9}
do
    port=900$i

    echo "Starting instance on port $port"

done

wait


network_error=${1:-false}

for i in {1..9}
do
    port=900$i
    port2=1000$i
    echo "go run . --grpc_server_address=:$port --http_server_address=:$port2 --network_error=$network_error"
    go run . --grpc_server_address=:$port --http_server_address=:$port2 --network_error=$network_error &
done

wait