FROM golang

WORKDIR /app
COPY ../../../go.mod ../../../go.sum ./
RUN go mod download

WORKDIR /app/examples/simplebank
COPY ../../go.mod ../../go.sum ./
RUN go mod download

WORKDIR /app
COPY . .
RUN go work sync

RUN go build ./examples/simplebank/cmd/consumer
ENTRYPOINT ./consumer
