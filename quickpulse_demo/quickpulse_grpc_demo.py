# QuickPulse gRPC Python Demo: Produce and Consume Example
#
# Prerequisites:
#   pip install grpcio grpcio-tools
#   python -m grpc_tools.protoc -I../proto --python_out=../quickpulse/proto --grpc_python_out=../quickpulse/proto ../proto/messagequeue.proto
#
# This will generate messagequeue_pb2.py and messagequeue_pb2_grpc.py in the quickpulse/proto directory.
#
# Run the quickpulse server in gRPC mode (default, port 50051) before running this script.

import sys
import os
sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), '..')))
import grpc
from quickpulse.proto import messagequeue_pb2
from quickpulse.proto import messagequeue_pb2_grpc

def main():
    channel = grpc.insecure_channel('localhost:50051')
    stub = messagequeue_pb2_grpc.MessageQueueStub(channel)

    # Produce a message
    message = b'Hello from Python gRPC client!'
    produce_req = messagequeue_pb2.ProduceRequest(payload=message)
    produce_resp = stub.Produce(produce_req)
    print("Produce response:", produce_resp.success, produce_resp.error)

    # Consume a message
    consume_req = messagequeue_pb2.ConsumeRequest()
    consume_resp = stub.Consume(consume_req)
    print("Consume response:", consume_resp.payload.decode('utf-8', errors='replace'), consume_resp.error)

if __name__ == '__main__':
    main()