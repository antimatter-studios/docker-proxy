#!/bin/bash -e

config=/etc/nginx/conf.d/default.conf
template="$(echo $1 | base64 -d)"

if [ ! -z "${template}" ]; then
    echo "Writing New Configuration"
    echo "${template}" > ${config}
else
    echo "Resetting Configuration back to default"
    cp /app/default.conf ${config}
fi

echo "Reloading NGINX"
nginx -s reload