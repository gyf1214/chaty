#!/bin/bash

rm -fr deploy/*
docker-build-go -path=. -cmd=cmd/server
chmod +x server
tar zcvf deploy/server.tar.gz conf/ server
tar zcvf deploy/static.tar.gz static/
rm server

# deploy to server
remote=$1
if [ -n "$remote" ]
then
    scp deploy/server.tar.gz $remote:~/chaty
    ssh $remote tar zxvf chaty/server.tar.gz -C chaty
fi
