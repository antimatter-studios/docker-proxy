#!/bin/bash
set -e

# generate the dhparam files needed for enabling SSL in the future
/app/scripts/generate-dhparam.sh

# copy the system resolvers to the nginx configuration
/app/scripts/resolvers.sh

# reset the configuration back to the default empty configuration
/app/scripts/reset.sh

# Ensure the socket directory exists
mkdir -p /var/run/proxy

# Start the management server in the background
echo "Starting management server..."
/usr/local/bin/management &

echo "Running '$@'"
exec "$@"
