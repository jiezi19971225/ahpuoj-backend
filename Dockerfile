FROM golang
ENV GOPROXY=https://goproxy.cn,direct
COPY ./ /backend
WORKDIR /backend
RUN go build

FROM ubuntu:18.04
LABEL authors="jiezi19971224@gmail.com"
COPY ./sources.list /etc/apt/sources.list

RUN set -ex \
    && export DEBIAN_FRONTEND=noninteractive \
    && apt-get clean \
    && apt-get update \
    && apt-get install -y tzdata

RUN set -ex \
    && apt-get clean \
    && ln -fs /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && dpkg-reconfigure -f noninteractive tzdata

WORKDIR /home/backend/

COPY ./config/auth_model.conf /home/backend/config/auth_model.conf
COPY ./config/config.ini.example /home/backend/config/config.ini

COPY --from=0 /backend/ahpuoj /home/backend/ahpuoj

ENV TINI_VERSION v0.19.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /tini
RUN chmod +x /tini
WORKDIR /home/backend/
ENTRYPOINT ["/tini", "--", "./ahpuoj"]