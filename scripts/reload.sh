#!/bin/bash -e

SRC=${CONFIG}/${NGINX_CONF#/}
DST=/etc/nginx/conf.d/default.conf

if [ -f "${SRC}" ]; then
    echo "Copying Custom NGINX Template configuration from '${SRC}' to '${DST}'"
    mv ${SRC} ${DST}
else
    echo "No Custom NGINX Template found..."
fi

echo "Reloading NGINX"
nginx -s reload