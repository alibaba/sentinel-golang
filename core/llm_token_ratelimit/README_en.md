#### Integration Steps

From the user's perspective, to integrate the Token rate limiting function provided by Sentinel, the following steps are required:

1. Prepare a Redis instance

2. Configure and initialize Sentinel's runtime environment.
   1. Only initialization from a YAML file is supported

3. Embed points (define resources) with fixed resource type: `ResourceType=ResTypeCommon` and `TrafficType=Inbound`

4. Load rules according to the configuration file below. The rule configuration items include: resource name, rate limiting strategy, specific rule items, Redis configuration, error code, and error message. The following is an example of rule configuration, with specific field meanings detailed in the "Configuration File Description" below.

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
   
5. Optional: Create an LLM instance and embed it into the provided adapter


#### Configuration Description

##### Configuration File

| Configuration Item | Type     | Required | Default Value                      | Description                                                  |
| :----------------- | :------- | :------- | :--------------------------------- | :----------------------------------------------------------- |
| enabled            | bool     | No       | false                              | Whether to enable the LLM Token Rate Limiting feature. Values: false (disable), true (enable) |
| redis              | object   | No       |                                    | Redis instance connection information                        |
| errorCode          | int      | No       | 429                                | Error code. If set to 0, it will be modified to 429 automatically |
| errorMessage       | string   | No       | "Too Many Requests"                | Error message                                                |

Redis Configuration

| Configuration Item | Type                 | Required | Default Value                            | Description                                                  |
| :----------------- | :------------------- | :------- | :--------------------------------------- | :----------------------------------------------------------- |
| addrs              | array of addr object | No       | [{name: "127.0.0.1", port: 6379}]        | Redis node service. **See Notes for details**                 |
| username           | string               | No       | Empty string                             | Redis username                                               |
| password           | string               | No       | Empty string                             | Redis password                                               |
| dialTimeout        | int                  | No       | 0                                        | Maximum waiting time for establishing a Redis connection, unit: milliseconds |
| readTimeout        | int                  | No       | 0                                        | Maximum waiting time for responses from the Redis server, unit: milliseconds |
| writeTimeout       | int                  | No       | 0                                        | Maximum time for sending command data to the network connection, unit: milliseconds |
| poolTimeout        | int                  | No       | 0                                        | Maximum waiting time to obtain an idle connection from the connection pool, unit: milliseconds |
| poolSize           | int                  | No       | 10                                       | Number of connections in the connection pool                  |
| minIdleConns       | int                  | No       | 5                                        | Minimum number of idle connections in the connection pool     |
| maxRetries         | int                  | No       | 3                                        | Maximum number of retries when an operation fails             |

Addr Configuration

| Configuration Item | Type   | Required | Default Value      | Description                                                  |
| :----------------- | :----- | :------- | :----------------- | :----------------------------------------------------------- |
| name               | string | No       | "127.0.0.1"        | Redis node service name. A complete [FQDN](https://en.wikipedia.org/wiki/Fully_qualified_domain_name) with service type, e.g., my-redis.dns, redis.my-ns.svc.cluster.local |
| port               | int    | No       | 6379               | Redis node service port                                      |

##### Rule Configuration

**Feature: Supports dynamic loading via LoadRules**

| Configuration Item | Type                         | Required | Default Value         | Description                                                  |
| :----------------- | :--------------------------- | :------- | :-------------------- | :----------------------------------------------------------- |
| resource           | string                       | No       | ".*"                  | Rule resource name, supports regular expressions. Values: ".*" (global match), user-defined regular expressions |
| strategy           | string                       | No       | "fixed-window"        | Rate limiting strategy. Values: fixed-window, peta (Prediction Error Temporal Allocation) |
| encoding           | object                       | No       |                       | Token encoding method. **Exclusive to PETA rate limiting strategy** |
| specificItems      | array of specificItem object | Yes      |                       | Specific rule items                                          |

encoding configuration

| Configuration Item | Type   | Required | Default Value | Description       |
| :----------------- | :----- | :------- | :------------ | :--------------- |
| provider           | string | No       | "openai"      | Model provider   |
| model              | string | No       | "gpt-4"       | Model name       |

specificItem configuration

| Configuration Item | Type                    | Required | Default Value | Description                                      |
| :----------------- | :---------------------- | :------- | :------------ | :----------------------------------------------- |
| identifier         | object                  | No       |               | Request identifier                               |
| keyItems           | array of keyItem object | Yes      |               | Key-value information for rule matching          |

identifier configuration

| Configuration Item | Type   | Required | Default Value | Description                                                  |
| :----------------- | :----- | :------- | :------------ | :----------------------------------------------------------- |
| type               | string | No       | "all"         | Request identifier type. Values: all (global rate limiting), header |
| value              | string | No       | ".*"          | Request identifier value, supports regular expressions. Values: ".*" (global match), user-defined regular expressions |

keyItem configuration

| Configuration Item | Type   | Required | Default Value | Description                                                  |
| :----------------- | :----- | :------- | :------------ | :----------------------------------------------------------- |
| key                | string | No       | ".*"          | Specific rule item value, supports regular expressions. Values: ".*" (global match), user-defined regular expressions |
| token              | object | Yes      |               | Token quantity and calculation strategy configuration       |
| time               | object | Yes      |               | Time unit and cycle configuration                           |

token configuration

| Configuration Item | Type   | Required | Default Value         | Description                                                  |
| :----------------- | :----- | :------- | :------------------- | :----------------------------------------------------------- |
| number             | int    | Yes      |                      | Token quantity, ≥ 0                                          |
| countStrategy      | string | No       | "total-tokens"       | Token calculation strategy. Values: input-tokens, output-tokens, total-tokens |

time configuration

| Configuration Item | Type   | Required | Default Value | Description                                                  |
| :----------------- | :----- | :------- | :------------ | :----------------------------------------------------------- |
| unit               | string | Yes      |               | Time unit. Values: second, minute, hour, day                  |
| value              | int    | Yes      |               | Time value, ≥ 0                                              |

#### Configuration File Example

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

#### LLM Framework Adaptation
Currently, non-intrusive integration of Sentinel's Token Rate Limiting capability is supported for the Langchaingo and Eino frameworks, mainly for text generation. For usage methods, please refer to:
- pkg/adapters/langchaingo/wrapper.go
- pkg/adapters/eino/wrapper.go

#### Notes

- Since only input tokens can be predicted currently, **it is recommended to use PETA for rate limiting input tokens**
- PETA uses tiktoken-go to estimate the number of input tokens consumed, but it is necessary to download or preconfigure the `Byte Pair Encoding (BPE)` dictionary:
  - Online Mode
    - When used for the first time, tiktoken-go needs to download the encoding file via the internet
  - Offline Mode
    - Prepare the pre-cached tiktoken-go encoding files (**not directly downloaded files, but files processed by tiktoken-go**) in advance, and specify the file directory by configuring the TIKTOKEN_CACHE_DIR environment variable
- Rule Deduplication Description
  - In keyItems, if only the "number" differs, duplicates will be removed and the latest "number" will be retained
  - In specificItems, only deduplicated keyItems will be retained
  - In resource, only the latest resource will be retained
- Redis Configuration Description
  - **If the connected Redis is in cluster mode, the number of addresses in "addrs" must be ≥ 2; otherwise, it will default to Redis single-node mode, causing rate limiting to fail**