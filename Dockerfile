# Build the manager binary
FROM golang:1.18 as builder

# Set the working directory
WORKDIR /workspace
# Copy Go module files and download dependencies
COPY go.mod go.sum /workspace/
RUN go mod download

# Copy the go source
COPY . /workspace/

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager main.go

# Use alpine as minimal base image to package the manager binary
FROM alpine:latest
WORKDIR /
COPY --from=builder /workspace/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
