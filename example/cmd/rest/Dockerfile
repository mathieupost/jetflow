FROM golang

WORKDIR /app
COPY ../../../go.mod ../../../go.sum ./
RUN go mod download

WORKDIR /app/example
COPY ../../go.mod ../../go.sum ./
RUN go mod download

WORKDIR /app
COPY . .
RUN go work sync

RUN go build ./example/cmd/rest
ENTRYPOINT ./rest
