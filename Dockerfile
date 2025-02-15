# Build
FROM golang:1.24-alpine AS builder
ENV GOFLAGS="-trimpath -mod=readonly -modcacherw"
ENV CGO_ENABLED=0
WORKDIR /go/src/app
RUN apk add --no-cache ca-certificates git libcap
RUN mkdir -p /opt/cache
COPY go.mod go.sum ./
RUN go mod download -x
COPY . .
RUN go build -v -tags prod -ldflags="-s -w" -o lancache
RUN setcap cap_net_bind_service=+ep /go/src/app/lancache
RUN update-ca-certificates

# Stage
FROM scratch AS root
COPY --chown=0:0 etc /etc
COPY --from=builder --chown=1234:1234 /opt/cache /opt/cache
COPY --chown=0:0 lancache.toml /opt/lancache.toml
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/app/lancache /bin/lancache

# Run
FROM scratch AS runner
COPY --from=root / /
USER app
WORKDIR /opt
ENV APP_ENV=production
VOLUME ["/opt/cache"]
ENTRYPOINT ["/bin/lancache"]
CMD [""]
