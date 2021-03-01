#!/usr/bin/bash

# files in assets with more then 100kb and count them
count=$(find web/assets -size '+100k' | wc -l)


# make sure count is an integer
let count=count+0


if [ $count -ne 0 ]; then
  echo "found $count big assets"
  exit 1
fi
