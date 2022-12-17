#!/bin/bash
set -e

# generate the dhparam files needed for enabling SSL in the future
/app/scripts/generate-dhparam.sh

# copy the system resolvers to the nginx configuration
/app/scripts/resolvers.sh

# reset the configuration back to the default empty configuration
/app/scripts/reset.sh

echo "Running '$@'"
exec "$@"