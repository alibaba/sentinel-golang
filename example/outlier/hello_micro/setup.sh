#!/bin/bash
network_error=${1:-false}

for i in {1..9}
do
    port=900$i
    echo "go run . --server_address=:$port --network_error=$network_error"
    go run . --server_address=:$port --network_error=$network_error &
done

wait