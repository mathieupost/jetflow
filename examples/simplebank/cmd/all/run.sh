#!/usr/bin/env sh

export CONSUMERS=10
export NATS_HOST=localhost
export JAEGER_HOST=hackbook

natsdata=$(mktemp -d)

# Function to kill all background processes when script exits
function cleanup {
    echo kill $natspid
    kill $natspid
    echo kill $restpid
    kill $restpid
    echo kill $conspid
    kill $conspid
    ssh hackbook "sleep 0.1; docker rm -f jaeger"
    rm -rf $natsdata
}
# Set trap to call cleanup on script exit
trap cleanup EXIT

# Start NATS JetStream in the background
go run github.com/nats-io/nats-server/v2 --jetstream --store_dir $natsdata &
natspid=$!
echo "NATS PID $natspid"

ssh hackbook "sleep 0.1; docker run -d --name jaeger -e COLLECTOR_OTLP_ENABLED=true -p 16686:16686 -p 4318:4318 jaegertracing/all-in-one:latest"

# Start all the consumers
seq 1 $CONSUMERS | xargs -n1 -I{} -P$CONSUMERS \
    env CONSUMER_ID={} \
    go run github.com/mathieupost/jetflow/examples/simplebank/cmd/consumer &
conspid=$!
echo "CONS PID $conspid"

# Start the rest server
go run github.com/mathieupost/jetflow/examples/simplebank/cmd/rest &
restpid=$!
echo "REST PID $restpid"

#!/bin/bash

# Set the URL you want to send requests to
URL="http://localhost:8080/health"

# Loop until a HTTP 200 response is received
while true; do
    RESPONSE=$(curl -o /dev/null -s -w "%{http_code}" "$URL")
    if [[ $RESPONSE -eq 200 ]]; then
        echo "Received HTTP 200 response."
        break  # Exit the loop if a HTTP 200 response is received
    else
        echo "Did not receive HTTP 200 response. Received HTTP $RESPONSE instead."
        sleep 1  # Wait for 1 seconds before sending the next request
    fi
done

wrk -t12 -c1000 -d1s http://localhost:8080/bench/$1

cleanup()
