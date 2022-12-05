#!/bin/bash
set -e

# generate the dhparam files needed for enabling SSL in the future
/app/scripts/generate-dhparam.sh

# copy the system resolvers to the nginx configuration
/app/scripts/resolvers.sh

# copy the template into the config directory, ready to be processed into an nginx configuration
/app/scripts/copy-template.sh

# Start the container with the default empty configuration
cp /app/default.conf /etc/nginx/conf.d/default.conf

echo "Running '$@'"
exec "$@"