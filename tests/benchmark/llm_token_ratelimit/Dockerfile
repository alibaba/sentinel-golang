FROM ubuntu:22.04


RUN sed -i 's/archive.ubuntu.com/mirrors.aliyun.com/g' /etc/apt/sources.list \
    && sed -i 's/security.ubuntu.com/mirrors.aliyun.com/g' /etc/apt/sources.list

RUN apt-get update && \
    apt-get install -y vim net-tools ca-certificates less redis-tools  && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /workspace
