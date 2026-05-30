# Stage 1: Build Go exporter
FROM docker.io/golang:1.23-alpine AS exporter-build

WORKDIR /build

# Copy only Go module files
COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o exporter ./cmd/exporter/

# Stage 2: Build supervise package
FROM docker.io/library/debian:trixie-slim AS supervise-build
ARG SUPERVISE_VERSION=8.9-2

RUN apt-get update && apt-get install -y curl

RUN mkdir /root/supervise
WORKDIR /root/supervise

RUN curl -o supervise.sh https://ragtech.com.br/Softwares_download/supervise-${SUPERVISE_VERSION}.sh
RUN chmod +x ./supervise.sh && ./supervise.sh --tar xpvf .
RUN dpkg-deb -xv Supervise.deb pkg
RUN mkdir /opt/supervise

# Copy only required files for supervise to work -- no fluff.
RUN for file in config.so device.so devices.xml monit.cfg monit.so supsvc web; do \
    cp -r /root/supervise/pkg/opt/supervise/$file /opt/supervise/; \
  done

# Stage 3: Runtime image
FROM docker.io/library/debian:trixie-slim

# supsvc only requires these packages below
RUN apt-get update && apt-get upgrade --yes && \
  apt-get install -y libqt5core5a libqt5script5 libqt5sql5 sqlite3 udev procps && \
  apt-get clean autoclean && apt-get autoremove --yes && rm -rf /var/lib/{apt,dpkg,cache,log}/

COPY --from=supervise-build /opt/supervise /opt/supervise
COPY --from=exporter-build /build/exporter /opt/exporter/exporter
COPY init.sh /init.sh

EXPOSE 4470 4471
VOLUME /data
ENTRYPOINT ["/init.sh"]