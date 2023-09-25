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
    ssh hackbook "sleep 0.1; docker rm -f jaeger" &
    rm -rf $natsdata
}
# Set trap to call cleanup on script exit
trap cleanup EXIT

# Start NATS JetStream in the background
go run github.com/nats-io/nats-server/v2 --jetstream --store_dir $natsdata &
natspid=$!
echo "NATS PID $natspid"

ssh hackbook "sleep 0.1; docker run -d --name jaeger -e COLLECTOR_OTLP_ENABLED=true -p 16686:16686 -p 4318:4318 jaegertracing/all-in-one:latest"

# Start the rest server
go run github.com/mathieupost/jetflow/examples/simplebank/cmd/rest &
restpid=$!
echo "REST PID $restpid"

# Start all the consumers
seq 1 $CONSUMERS | xargs -n1 -I{} -P$CONSUMERS \
    env CONSUMER_ID={} \
    go run github.com/mathieupost/jetflow/examples/simplebank/cmd/consumer
