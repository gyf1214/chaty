#!/bin/bash

# pack
deploy=deploy
server=server.tar.gz
serverlist="conf/server.json server ctl.sh"
cmd=server

mkdir -p $deploy
rm $deploy/$server
docker-build-go -path=. -cmd=cmd/$cmd
chmod +x $cmd
tar zcvf $deploy/$server $serverlist
rm $cmd

# deploy to server
remote=$1
path="~/chaty"

read -r -d '' cmd <<- EOF
    cd ${path} && pwd &&
    tar zxvf ${server} &&
    ./ctl.sh restart &&
    echo finished
EOF

if [ -n "$remote" ]; then
    scp $deploy/$server $remote:$path
    echo $cmd
    ssh $remote bash -c \"$cmd\"
fi
