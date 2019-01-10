# Dockerfile
FROM registry.cn-hangzhou.aliyuncs.com/basicimage/golang:latest
MAINTAINER Jhhe <hejianhua@wonlycloud.com>

# copy
COPY supervisord.conf /etc/supervisord.conf
COPY gopath/ /tmp/gopath/
COPY das/ /tmp/das/

# 设置时区
RUN /bin/cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo 'Asia/Shanghai' >/etc/timezone \
	
# install
RUN set -x \
    && cd /tmp \
    && chmod 755 -R gopath \
    && cp -rfp /tmp/gopath/ /www/wonly/gopath/ \
    && chmod 755 -R das && cd das \
	&& cp -rfp /tmp/das/src/das.ini /www/wonly/DAS_go/ \
    && go build -o das/bin/das das/src/.  \
    && mkdir -p /www/wonly/DAS_go \
    && mkdir -p /www/wonly/DAS_go/logs \
    && cp -rfp /tmp/das/bin/das /www/wonly/DAS_go/ \
    && chown nobody:nobody -R /www/wonly/DAS_go \
    && rm -rf /tmp/*

WORKDIR /www/wonly/DAS_go
VOLUME ["/www/wonly/DAS_go/logs"]

EXPOSE 10700
CMD ["supervisord"]
