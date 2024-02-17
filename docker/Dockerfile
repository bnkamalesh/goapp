FROM golang:1.21 AS builder

COPY ../ /app
WORKDIR /app
RUN ls
# Toggle CGO on your app requirement
RUN CGO_ENABLED=0 go build -ldflags '-s -w -extldflags "-static"' -o /app/appbin main.go
# Use below if using vendor
# RUN CGO_ENABLED=0 go build -mod=vendor -ldflags '-extldflags "-static"' -o /app/appbin *.go

FROM debian:stable-slim
LABEL MAINTAINER Author <author@example.com>

# Following commands are for installing CA certs (for proper functioning of HTTPS and other TLS)
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates  \
    netbase \
    && rm -rf /var/lib/apt/lists/ \
    && apt-get autoremove -y && apt-get autoclean -y

# Add new user 'appuser'. App should be run without root privileges as a security measure
RUN adduser --home "/appuser" --disabled-password appuser \
    --gecos "appuser,-,-,-"
USER appuser

COPY --from=builder /app/internal/server/http/web /home/appuser/app/web
COPY --from=builder /app/appbin /home/appuser/app

ENV TEMPLATES_BASEPATH=/home/appuser/app/web/templates

WORKDIR /home/appuser/app

# Since running as a non-root user, port bindings < 1024 are not possible
# 8000 for HTTP; 8443 for HTTPS;
EXPOSE 8000

CMD ["./appbin"]