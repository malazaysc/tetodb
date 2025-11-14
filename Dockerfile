FROM golang:1.23-alpine

WORKDIR /build

# Copy Go module files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY engine/ ./engine/
COPY wasm/ ./wasm/

# Create output directory
RUN mkdir -p /output

# Build the WASM module
RUN GOOS=js GOARCH=wasm go build -o /output/tetodb.wasm ./wasm

# Copy the Go WASM runtime
RUN cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" /output/

# Set the output directory as the working directory
WORKDIR /output

CMD ["sh", "-c", "echo 'Build complete! Files:' && ls -lh"]
