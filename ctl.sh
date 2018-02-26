#!/bin/bash

binfile=./server
logfile=server.log
lockfile=server.pid

start() {
    if [ -f $lockfile ]; then
        echo 'already running'
    else
        echo 'start server...'
        $binfile >> $logfile 2>&1 < /dev/null & disown
        echo $! > $lockfile
        echo 'finished'
    fi
}

stop() {
    if [ ! -f $lockfile ]; then
        echo 'not running'
    else
        echo 'stop server...'
        pid=`cat $lockfile`
        kill $pid
        echo 'wait for server to exit...'
        while ps -p $pid > /dev/null; do sleep 1; done
        rm $lockfile
        echo 'finished'
    fi
}

case $1 in
start)
    start
    ;;
stop)
    stop
    ;;
restart)
    stop
    start
    ;;
*)
    echo './ctl.sh {start | stop | restart}'
    exit -1
    ;;
esac
