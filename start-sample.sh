#!/bin/sh

# This is sample I start my app
# Run the binary with option -h for all help
# You need to setup the db first? Run
# ./webnote-bin -db path-to-your-new-db.db -key path-to-ssl-key -cert path-to-cert-file -p port -setup

export TZ=Australia/Brisbane
export TMPDIR=/mnt/data/tmp

nohup ./webnote-go -db testwebnote.db -key yoursite.key -cert yoursite.crt -p 443 -baseurl https://your-site-domain:443 &
