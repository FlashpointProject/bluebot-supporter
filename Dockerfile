FROM golang:1.27.4-trixie@sha256:f18c3d73a184fb0582828510722e723b63a93f7188d11dd223d9be6f6880e65d AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -trimpath -o /out/bluebot .

FROM debian:trixie-20261104-slim@sha256:a72b4433de1cf7d6bae0c4cee74d0547f58601a643b3375f070c51600d63b6b2
RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates \
 && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=build /out/bluebot /bluebot
# The database is not part of the image: mount it at runtime.
#   /data/master.db      (required; only read, by /tags and /launch-commands)
#   /app/config.json     (optional; see config.example.json, defaults to port 80)
ENV BIND_ADDR=0.0.0.0 \
    DB_PATH=/data/master.db
VOLUME /data
CMD ["/bluebot"]
