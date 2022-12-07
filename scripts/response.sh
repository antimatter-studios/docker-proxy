#!/bin/bash -e

echo "Writing Configuration"
echo "$(echo $1 | base64 -d)" > /etc/nginx/conf.d/default.conf

echo "Reloading NGINX"
nginx -s reload