#!/bin/bash

URL=https://www.dropbox.com/s/duv704waqjp3tu1/hn_logs.tsv.gz?dl=1
FILE=hn_logs.tsv

if [ -f $FILE ]; then
  exit 0
fi

wget -O $FILE.gz $URL
gzip -f -d $FILE.gz