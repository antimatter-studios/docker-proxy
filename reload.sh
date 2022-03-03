#!/bin/bash -e

echo "Adding Custom NGINX Template configuration"
mv ${NGINX_CONF} /etc/nginx/conf.d/default.conf

echo "Reloading NGINX"
nginx -s reload