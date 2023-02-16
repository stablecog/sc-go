# Should be built from root directory of the workspace
FROM golang:1.20-buster as builder

WORKDIR /app

ADD . .
RUN cd server && go build -ldflags '-s -w' -o /app/sc-go-server

FROM debian:buster-slim

RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
  ca-certificates && \
  rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/sc-go-server /app/server

EXPOSE 13337

CMD ["/app/server"]