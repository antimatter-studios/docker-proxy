#!/bin/bash -e

SRC=/app/${TEMPLATE}
DIR=$(dirname ${CONFIG}/${NGINX_CONF#/})
DST=${DIR}/${TEMPLATE}

if [ -f "${DIR}" ]; then 
	echo "Error, directory '${DIR}' was a file, deleting it"
	rm ${DIR}
fi

if [ ! -d "${DIR}" ]; then
	echo "Directory '${DIR}' does not exist, creating it"
	mkdir -p ${DIR}
fi

if [ ! -f "${DST}" ]; then
	echo "Copying '${SRC}' template to '${DST}'"
	cp ${SRC} ${DST}
else
	echo "Overwriting contents of '${DST}' with contents of '${SRC}' template"
	cat ${SRC} > ${DST}
fi