FROM registry.cn-hangzhou.aliyuncs.com/basicimage/golang:latest AS builder
WORKDIR /tmp/
COPY ./das/ /tmp/das/

RUN go env -w GOPROXY=https://goproxy.cn,direct \
    && go env -w GO111MODULE=on \
    && cd /tmp/das/src && ls -l && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -v -o ../bin/das.bin

FROM alpine:3.7
WORKDIR /www/wonly/DAS_go/
COPY ./das/src/cert/ /www/wonly/DAS_go/cert/
COPY --from=builder /tmp/das/bin/das.bin /tmp/das/src/*.ini /www/wonly/DAS_go/
COPY supervisord.conf /etc/supervisord.conf

RUN echo "http://mirrors.aliyun.com/alpine/v3.7/main/" > /etc/apk/repositories \
    && apk add tzdata && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone \
    && apk add supervisor \
    && apk del tzdata

VOLUME ["/www/wonly/DAS_go/logs"]

EXPOSE 10701
EXPOSE 10702
EXPOSE 10703
EXPOSE 10704

CMD ["supervisord"]