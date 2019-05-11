#!/usr/bin/env bash
echo '获取包...'
go get -u github.com/gorilla/mux
echo '...获取完毕'
echo '正在编译...' && go build && echo '...编译成功' && pwd && ls -lh && ./app