#!/bin/bash

WORKERS=20
TOTAL_MESSAGES=100000
DURATION=60
START=$(date +%s)
PROTO_DIR="proto"
PROTO_FILE="messagequeue.proto"
PAYLOAD='{"payload": "YXNkYXNka2poYXNrZGpoYXNrZGhqYWtzZGhha3NqaGRrYWpzaGQ="}'
GRPC_SERVER="localhost:50051"
GRPC_METHOD="messagequeue.MessageQueue/Produce"

COUNTER_FILE=$(mktemp)
echo 0 > "$COUNTER_FILE"

echo "Starting perf test: $TOTAL_MESSAGES messages, $WORKERS workers, $DURATION seconds max"
echo "gRPC server: $GRPC_SERVER, method: $GRPC_METHOD"
echo "Start time: $(date)"

worker() {
  local wid=$1
  echo "Worker $wid started"
  while true; do
    NOW=$(date +%s)
    if (( NOW - START >= DURATION )); then
      echo "Worker $wid exiting (time limit reached)"
      break
    fi
    COUNT=$(perl -e 'open F, "+<", $ARGV[0]; flock F, 2; $c = <F>; seek F, 0, 0; $c++; print F $c; truncate F, tell(F); print $c;' "$COUNTER_FILE")
    if (( COUNT > TOTAL_MESSAGES )); then
      echo "Worker $wid exiting (message limit reached)"
      break
    fi
    grpcurl -plaintext -import-path "$PROTO_DIR" -proto "$PROTO_FILE" \
      -d "$PAYLOAD" "$GRPC_SERVER" "$GRPC_METHOD" >/dev/null 2>&1
    # Log every 10,000 messages
    if (( COUNT % 10000 == 0 )); then
      echo "Progress: $COUNT messages published (by worker $wid, $(date))"
    fi
  done
}

# Progress logger
progress_logger() {
  while true; do
    sleep 10
    COUNT=$(cat "$COUNTER_FILE")
    echo "Progress: $COUNT messages published so far ($(date))"
    NOW=$(date +%s)
    if (( NOW - START >= DURATION )); then
      break
    fi
    if (( COUNT >= TOTAL_MESSAGES )); then
      break
    fi
  done
}

for ((i=0; i<WORKERS; i++)); do
  worker $i &
done

progress_logger

wait

END=$(date +%s)
ELAPSED=$((END - START))
SUCCESS=$(cat "$COUNTER_FILE")
if (( SUCCESS > TOTAL_MESSAGES )); then
  SUCCESS=$TOTAL_MESSAGES
fi

echo "Test complete!"
echo "Total messages published: $SUCCESS"
echo "Elapsed time: $ELAPSED seconds"
echo "Throughput: $((SUCCESS / ELAPSED)) messages/sec"
echo "End time: $(date)"

rm "$COUNTER_FILE"