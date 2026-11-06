FROM golang:1.27.3-trixie@sha256:af1bfc2614ebb1dbd31af514b78740fd678850e76498762b425ecee4c5c6ea80 AS build
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
