FROM golang:1.18 AS build

WORKDIR /go/src

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY pkg/instaclustr/mock/server/main.go pkg/instaclustr/mock/server/main.go
COPY pkg/instaclustr/mock/server/go pkg/instaclustr/mock/server/go

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o openapi ./pkg/instaclustr/mock/server/main.go

FROM scratch AS runtime
COPY --from=build /go/src/openapi ./
EXPOSE 8090/tcp
ENTRYPOINT ["./openapi"]
