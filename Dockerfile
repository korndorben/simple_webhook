FROM golang:1.8

WORKDIR /go/src/app

# 端口
EXPOSE 8080

# 开放容器内的目录
VOLUME ["/go/src/app"]

CMD ["/bin/bash", "./run.sh"]
