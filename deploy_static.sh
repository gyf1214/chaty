#!/bin/bash

# pack
deploy=deploy
static=static.tar.gz
staticlist=static/

mkdir -p $deploy
rm $deploy/$static
tar zcvf $deploy/$static $staticlist

# deploy to web server
remote=$1
path=/var/www/chaty
restart="service nginx restart"

read -r -d '' cmd <<- EOF
    pwd && sudo cp ${static} ${path}/ &&
    cd ${path} && pwd &&
    sudo tar zxvf ${static} &&
    sudo ${restart} &&
    echo finished
EOF

if [ -n "$remote" ]; then
    scp $deploy/$static $remote:~
    echo $cmd
    ssh -t $remote bash -c \"$cmd\"
fi
