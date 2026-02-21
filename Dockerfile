FROM golang:1.24-alpine AS builder
WORKDIR /build
COPY go.mod ./
RUN go mod download
COPY cmd/ cmd/
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /management ./cmd/management

FROM nginx:alpine
LABEL maintainer="Chris Thomas <chris.alex.thomas@gmail.com> (@chrisalexthomas)"

# Install wget and install/updates certificates
RUN apk add --no-cache --virtual .run-deps \
    ca-certificates bash wget openssl \
    && update-ca-certificates

# Configure Nginx and apply fix for very long server names
RUN echo "Writing custom nginx configuration values" \
 && sed -i 's/worker_processes  1/worker_processes  auto/' /etc/nginx/nginx.conf \
 && sed -i 's/worker_connections  1024/worker_connections  10240/' /etc/nginx/nginx.conf

COPY network_internal.conf /etc/nginx/
COPY proxy.conf /etc/nginx/
COPY html /etc/nginx/html

COPY . /app
WORKDIR /app/

COPY --from=builder /management /usr/local/bin/management

COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

RUN chmod +x /app/scripts/*.sh

VOLUME ["/etc/nginx/certs", "/etc/nginx/dhparam", "/var/run/proxy"]

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["nginx", "-g", "daemon off;"]
