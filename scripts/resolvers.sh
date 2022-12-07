#!/bin/bash -e

# Compute the DNS resolvers for use in the templates - if the IP contains ":", it's IPv6 and must be enclosed in []
RESOLVERS=$(awk '$1 == "nameserver" {print ($2 ~ ":")? "["$2"]": $2}' ORS=' ' /etc/resolv.conf | sed 's/ *$//g');

SCOPED_IPV6_REGEX="\[fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}\]"

if [ -z "${RESOLVERS}" ]; then
	echo "Warning: unable to determine DNS resolvers for nginx"
	unset RESOLVERS
elif [[ ${RESOLVERS} =~ ${SCOPED_IPV6_REGEX} ]]; then
	echo -n "Warning: Scoped IPv6 addresses removed from resolvers: \c"
	echo "${RESOLVERS}" | grep -Eo "${SCOPED_IPV6_REGEX}" | paste -s -d ' '
	RESOLVERS=$(echo "${RESOLVERS}" | sed -r "s/${SCOPED_IPV6_REGEX}//g" | xargs echo -n); 
fi

if [ ! -z "${RESOLVERS}" ]; then
	echo "Writing /etc/nginx/conf.d/resolvers.conf"
	echo "resolver ${RESOLVERS};" > /etc/nginx/conf.d/resolvers.conf
fi