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

COPY . /app/
WORKDIR /app/

VOLUME ["/etc/nginx/certs", "/etc/nginx/dhparam"]

ENTRYPOINT ["/app/docker-entrypoint.sh"]
CMD ["nginx", "-g", "daemon off;"]
