FROM nginx:alpine
LABEL maintainer="Chris Thomas <chris.alex.thomas@gmail.com> (@chrisalexthomas)"

# These are defaults that you can override in your docker-compose if you want
ENV CONFIG=/config
ENV NGINX_CONF=docker-proxy/nginx.conf
ENV TEMPLATE=nginx.template

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

COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

RUN chmod +x /app/scripts/*.sh

VOLUME ["/etc/nginx/certs", "/etc/nginx/dhparam"]

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["nginx", "-g", "daemon off;"]
