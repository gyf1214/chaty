#!/bin/bash

rm -fr deploy/*
docker-build-go -path=. -cmd=cmd/server
tar zcvf deploy/server.tar.gz conf/ server
tar zcvf deploy/static.tar.gz static/
rm server
