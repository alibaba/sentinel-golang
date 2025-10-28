#### 接入步骤

从用户角度，接入Sentinel提供的Token限流功能，需要以下几步：

1. 准备Redis实例

2. 对 Sentinel 的运行环境进行相关配置并初始化。

   1. 仅支持从yaml文件初始化

3. 埋点（定义资源），固定`ResourceType=ResTypeCommon`且`TrafficType=Inbound`的资源类型

4. 根据下面的配置文件加载规则，规则配置项包括：资源名称、限流策略、具体规则项、redis配置、错误码、错误信息。如下是配置规则的示例，具体字段含义在下文的“配置文件描述”中有具体说明。

   ```go
   _, err = llmtokenratelimit.LoadRules([]*llmtokenratelimit.Rule{
       {
   
           Resource: ".*",
           Strategy: llmtokenratelimit.FixedWindow,
           SpecificItems: []llmtokenratelimit.SpecificItem{
               {
                   Identifier: llmtokenratelimit.Identifier{
                       Type:  llmtokenratelimit.Header,
                       Value: ".*",
                   },
                   KeyItems: []llmtokenratelimit.KeyItem{
                       {
                           Key:      ".*",
                           Token: llmtokenratelimit.Token{
                               Number:        1000,
                               CountStrategy: llmtokenratelimit.TotalTokens,
                           },
                           Time: llmtokenratelimit.Time{
                               Unit:  llmtokenratelimit.Second,
                               Value: 60,
                           },
                       },
                   },
               },
           },
       },
   })
   ```
   
5. 可选：创建LLM实例嵌入到提供的适配器中即可

#### 配置描述

##### 配置文件

| 配置项       | 类型   | 必填 | 默认值              | 说明                                                       |
| :----------- | :----- | :--- | :------------------ | :--------------------------------------------------------- |
| enabled      | bool   | 否   | false               | 是否启用LLM Token限流功能，取值：false(不启用)、true(启用) |
| redis        | object | 否   |                     | redis实例连接信息                                          |
| errorCode    | int    | 否   | 429                 | 错误码，设置为0时会修改为429                               |
| errorMessage | string | 否   | "Too Many Requests" | 错误信息                                                   |

redis配置

| 配置项       | 类型                 | 必填 | 默认值                            | 说明                                               |
| :----------- | :------------------- | :--- | :-------------------------------- | :------------------------------------------------- |
| addrs        | array of addr object | 否   | [{name: "127.0.0.1", port: 6379}] | redis节点服务，**见注意事项说明**                  |
| username     | string               | 否   | 空字符串                          | redis用户名                                        |
| password     | string               | 否   | 空字符串                          | redis密码                                          |
| dialTimeout  | int                  | 否   | 0                                 | 建立redis连接的最长等待时间，单位：毫秒            |
| readTimeout  | int                  | 否   | 0                                 | 等待Redis服务器响应的最长时间，单位：毫秒          |
| writeTimeout | int                  | 否   | 0                                 | 向网络连接发送命令数据的最长时间，单位：毫秒       |
| poolTimeout  | int                  | 否   | 0                                 | 从连接池获取一个空闲连接的最大等待时间，单位：毫秒 |
| poolSize     | int                  | 否   | 10                                | 连接池中的连接数量                                 |
| minIdleConns | int                  | 否   | 5                                 | 连接池闲置连接的最少数量                           |
| maxRetries   | int                  | 否   | 3                                 | 操作失败，最大尝试次数                             |

addr配置

| 配置项 | 类型   | 必填 | 默认值      | 说明                                                         |
| :----- | :----- | :--- | :---------- | :----------------------------------------------------------- |
| name   | string | 否   | "127.0.0.1" | redis节点服务名称，带服务类型的完整 [FQDN](https://en.wikipedia.org/wiki/Fully_qualified_domain_name) 名称，例如 my-redis.dns、redis.my-ns.svc.cluster.local |
| port   | int    | 否   | 6379        | redis节点服务端口                                            |

##### 规则配置

**特点：支持LoadRules动态加载**

| 配置项        | 类型                         | 必填 | 默认值         | 说明                                                         |
| :------------ | :--------------------------- | :--- | :------------- | :----------------------------------------------------------- |
| resource      | string                       | 否   | ".*"           | 规则资源名称，支持正则表达式，取值：".*"(全局匹配)、用户自定义正则表达式 |
| strategy      | string                       | 否   | "fixed-window" | 限流策略，取值：fixed-window（固定窗口）、peta（预测误差时序分摊） |
| encoding      | object                       | 否   |                | token编码方式，**专用于peta限流策略**                        |
| specificItems | array of specificItem object | 是   |                | 具体规则项                                                   |

encoding配置

| 配置项   | 类型   | 必填 | 默认值   | 说明     |
| :------- | :----- | :--- | :------- | :------- |
| provider | string | 否   | "openai" | 模型厂商 |
| model    | string | 否   | "gpt-4"  | 模型名称 |

specificItem配置

| 配置项     | 类型                    | 必填 | 默认值 | 说明               |
| :--------- | :---------------------- | :--- | :----- | :----------------- |
| identifier | object                  | 否   |        | 请求标识符         |
| keyItems   | array of keyItem object | 是   |        | 规则匹配的键值信息 |

identifier配置

| 配置项 | 类型   | 必填 | 默认值 | 说明                                                         |
| :----- | :----- | :--- | :----- | :----------------------------------------------------------- |
| type   | string | 否   | "all"  | 请求标识符类型，取值：all(全局限流)、header                  |
| value  | string | 否   | ".*"   | 请求标识符取值，支持正则表达式，取值：".*"(全局匹配)、用户自定义正则表达式 |

keyItem配置

| 配置项 | 类型   | 必填 | 默认值 | 说明                                                         |
| :----- | :----- | :--- | :----- | :----------------------------------------------------------- |
| key    | string | 否   | ".*"   | 具体规则项取值，支持正则表达式，取值：".*"(全局匹配)、用户自定义正则表达式 |
| token  | object | 是   |        | token数量和计算策略配置                                      |
| time   | object | 是   |        | 时间单位和周期配置                                           |

token配置

| 配置项        | 类型   | 必填 | 默认值         | 说明                                                         |
| :------------ | :----- | :--- | :------------- | :----------------------------------------------------------- |
| number        | int    | 是   |                | token数量，大于等于0                                         |
| countStrategy | string | 否   | "total-tokens" | token计算策略，取值：input-tokens、output-tokens、total-tokens |

time配置

| 配置项 | 类型   | 必填 | 默认值 | 说明                                      |
| :----- | :----- | :--- | :----- | :---------------------------------------- |
| unit   | string | 是   |        | 时间单位，取值：second、minute、hour、day |
| value  | int    | 是   |        | 时间值，大于等于0                         |

#### 配置文件示例

```YAML
version: "v1"
sentinel:
  app:
    name: sentinel-go-demo
  log:
    metric:
      maxFileCount: 7
  llmTokenRatelimit:
  	enabled: true
  	
    errorCode: 429
    errorMessage: "Too Many Requests"
    
    redis:
      addrs:
        - name: "127.0.0.1"
          port: 6379
      username: "redis"
      password: "redis"
      dialTimeout: 5000
      readTimeout: 5000
      writeTimeout: 5000
      poolTimeout: 5000
      poolSize: 10
      minIdleConns: 5
      maxRetries: 3
```
#### LLM框架适配

目前支持Langchaingo和Eino框架无侵入式接入Sentinel提供的Token限流能力，主要适用于文本生成方面，使用方法详见：

- pkg/adapters/langchaingo/wrapper.go
- pkg/adapters/eino/wrapper.go

#### 注意事项

- 由于目前仅可预知input tokens，所以**建议使用PETA专对于input tokens进行限流**
- PETA使用tiktoken-go预估输入消耗token数，但是需要下载或预先配置`字节对编码(Byte Pair Encoding，BPE)`字典
  - 在线模式
    - 首次使用时，tiktoken-go需要联网下载编码文件
  - 离线模式
    - 预先准备缓存好的tiktoken-go的编码文件（**非直接下载文件，而是经过tiktoken-go处理后的文件**），并通过配置TIKTOKEN_CACHE_DIR环境变量指定文件目录位置
- 规则去重说明
  - keyItems中，若仅number不同，会去重保留最新的number
  - specificItems中，仅保留去重后的keyItems
  - resource中，仅保留最新的resource
- redis配置说明
  - **若连接的redis是集群模式，那么addrs里面的地址数量必须大于等于2个，否则会默认进入redis单点模式，导致限流失效**
