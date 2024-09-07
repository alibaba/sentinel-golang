# Quick Start

## 概述

目的：构建用于测试离群摘除功能的随机失败环境

方法：服务由10个节点实例构成，每个节点编号id，启动后从10s+id*5s时开始出现异常，时间段持续10s

- 为模拟网络错误，每个节点在10s后阻塞服务调用，直到第15s才恢复
- 为模拟业务错误，每个节点在10s后开始返回500错误，直到第15s才恢复

## 使用

1、安装注册中心（可选）

安装etcd：https://etcd.io/docs/v3.5/install/

2、启动注册中心

```bash
etcd 
```

etcd默认端口为2379

3、启动9个server进程

go-micro框架：
```
cd helloworld && ./setup.sh
```
kitex框架：
```
cd hellokitex && ./setup.sh
```
kratos框架：
```
cd hellokratos && ./setup.sh
```

4、客户端测试`helloworld/client/client_test.go`