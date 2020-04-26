#!/bin/sh

# sample to extract and deploy
killall webnote-bin
tar xf $1 -C /data/scripts
( cd /data/scripts/webnote-go-bin/ ; ./start.sh )
