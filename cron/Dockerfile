# Should be built from root directory of the workspace
FROM golang:1.20-buster as builder

WORKDIR /app

ADD . .
RUN cd cron && go build -ldflags '-s -w' -o /app/sc-cron

FROM debian:buster-slim

RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
  ca-certificates && \
  rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/sc-cron /app/cron

CMD ["/app/cron", "-all"]