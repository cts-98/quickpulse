# Makefile for quickpulse: Build and run the message queue server in different modes

# Build the quickpulse server binary
build:
	go build -o quickpulse-server .

# Run in gRPC unary mode (default port: 50051)
run-grpc: build
	@echo '***************************************************'
	@echo '*   ____       _      _      ___       _          *'
	@echo '*  /___ \_   _(_) ___| | __ / _ \_   _| |___  ___ *'
	@echo '* //  / / | | | |/ __| |/ // /_)/ | | | / __|/ _ \*'
	@echo '*/ \_/ /| |_| | | (__|   </ ___/| |_| | \__ \  __/*'
	@echo '*\___,_\ \__,_|_|\___|_|\_\/     \__,_|_|___/\___|*'
	@echo '***************************************************'
	@printf '⚡⚡⚡ \033[1mBlazing fast, in-memory message queue\033[0m ⚡⚡⚡\n'
	WS_MODE=0 RPC_MODE=1 RPC_STREAM_MODE=0 ./quickpulse-server

# Run in gRPC streaming mode (default port: 50051)
run-grpc-stream: build
	@echo '***************************************************'
	@echo '*   ____       _      _      ___       _          *'
	@echo '*  /___ \_   _(_) ___| | __ / _ \_   _| |___  ___ *'
	@echo '* //  / / | | | |/ __| |/ // /_)/ | | | / __|/ _ \*'
	@echo '*/ \_/ /| |_| | | (__|   </ ___/| |_| | \__ \  __/*'
	@echo '*\___,_\ \__,_|_|\___|_|\_\/     \__,_|_|___/\___|*'
	@echo '***************************************************'
	@printf '⚡⚡⚡ \033[1mBlazing fast, in-memory message queue\033[0m ⚡⚡⚡\n'
	WS_MODE=0 RPC_MODE=0 RPC_STREAM_MODE=1 ./quickpulse-server

# Run in WebSocket mode (default port: 8081)
run-ws: build
	@echo '***************************************************'
	@echo '*   ____       _      _      ___       _          *'
	@echo '*  /___ \_   _(_) ___| | __ / _ \_   _| |___  ___ *'
	@echo '* //  / / | | | |/ __| |/ // /_)/ | | | / __|/ _ \*'
	@echo '*/ \_/ /| |_| | | (__|   </ ___/| |_| | \__ \  __/*'
	@echo '*\___,_\ \__,_|_|\___|_|\_\/     \__,_|_|___/\___|*'
	@echo '***************************************************'
	@printf '⚡⚡⚡ \033[1mBlazing fast, in-memory message queue\033[0m ⚡⚡⚡\n'
	WS_MODE=1 RPC_MODE=0 RPC_STREAM_MODE=0 ./quickpulse-server

# Clean build artifacts
clean:
	rm -f quickpulse-server

# Usage instructions
help:
	@echo "Usage:"
	@echo "  make build            # Build the quickpulse server binary (quickpulse-server)"
	@echo "  make run-grpc         # Run in gRPC unary mode (port 50051)"
	@echo "  make run-grpc-stream  # Run in gRPC streaming mode (port 50051)"
	@echo "  make run-ws           # Run in WebSocket mode (port 8081)"
	@echo "  make clean            # Remove the quickpulse binary"
	@echo ""
	@echo "Prometheus metrics are available at http://localhost:8080/metrics"
	@echo "Only one mode can be active at a time. Set the appropriate environment variable."

.PHONY: build run-grpc run-grpc-stream run-ws clean help