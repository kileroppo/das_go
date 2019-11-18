# Dockerfile
FROM registry.cn-hangzhou.aliyuncs.com/basicimage/golang:latest
MAINTAINER Jhhe <hejianhua@wonlycloud.com>

# copy
COPY supervisord.conf /etc/supervisord.conf
COPY ./gopath/ /usr/local/gopath/
COPY ./das/ /tmp/das/

# 设置时区
RUN /bin/cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo 'Asia/Shanghai' >/etc/timezone
	
# install
RUN set -x \
    && cd /tmp \
    && cd /usr/local && ls -l \
    && mkdir -p /www/wonly/DAS_go \
    && mkdir -p /www/wonly/DAS_go/logs

RUN cd /tmp && ls -l
RUN cd /tmp/das && ls -l && cp /tmp/das/src/das.ini /www/wonly/DAS_go/ && cp /tmp/das/src/das_dev.ini /www/wonly/DAS_go/
RUN cd /www/wonly/DAS_go/ && ls -l
RUN cd /tmp/das/src && ls -l && GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -v -o ../bin/das.bin .
RUN cp /tmp/das/bin/das.bin /www/wonly/DAS_go/
RUN cp -rf /tmp/das/src/cert /www/wonly/DAS_go/.
RUN chmod -R 755 /www/wonly/DAS_go
RUN cd /www/wonly/DAS_go/ && ls -l
RUN rm -rf /tmp/*

WORKDIR /www/wonly/DAS_go
VOLUME ["/www/wonly/DAS_go/logs"]

EXPOSE 10701
EXPOSE 10702
EXPOSE 10703
EXPOSE 10704
EXPOSE 6060

CMD ["supervisord"]
