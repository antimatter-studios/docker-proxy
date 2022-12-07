#!/bin/bash
set -e

# generate the dhparam files needed for enabling SSL in the future
/app/scripts/generate-dhparam.sh

# copy the system resolvers to the nginx configuration
/app/scripts/resolvers.sh

# Start the container with the default empty configuration
cp /app/default.conf /etc/nginx/conf.d/default.conf

echo "Running '$@'"
exec "$@"