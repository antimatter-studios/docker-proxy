#!/bin/bash
set -e

/app/scripts/generate-dhparam.sh
/app/scripts/resolvers.sh
/app/scripts/copy-template.sh

TEST=${CONFIG}/${NGINX_CONF#/}

if [ ! -f "${TEST}" ]; then
    echo "We could not find the file '${TEST}', there must be an error in the configuration, exiting..."
    exit 1
fi

echo "Running '$@'"
exec "$@"