#!/bin/bash

case $1 in
	-h | --help ) echo "usage: $(basename $0) DIR"; exit;;
esac

if [ $# -ne 1 ]; then
	echo "error: wrong number of arguments" 1>&2
	exit 1
fi

go build -o logs -trimpath ./$1
