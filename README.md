# anygo
一个 * 极少依赖 * 的 Go RPC 框架和基础库。

RPC Client 功能：
- 下游服务管理器：抽象的下游服务 (Service)，可通过配置文件管理下游的信息，如连接超时、socket 读写超时、重试、
  网络连接池、TLS 认证配置、代理等信息，支持多种服务发现方式(已内置DNS、File、Host+Port，支持自定义)、
  支持多种负载均衡策略（已内置 Random、RoundRobin，支持自定义）
- DNS 解析组件：内置缓存功能、内置拦截器功能
- TCP 拨号组件：内置拦截器功能
- 通用的 TCP RPC Client 组件，内置拦截器功能，已实现协议：
  1. Redis： Redis Resp3 协议的 Client
  2. HTTP：HTTP/HTTPS 的 Client
- Redis Client ( store/xredis ): 使用 resp3 协议的 Redis Client
- db Client (store/xdb )：内置拦截器功能，轻量 ORM 支持，内置已支持 MariaDB、MySQL、SQL Server、Sqlite、Postgres
  - 注：需要自己注册对应数据库的驱动 

RPC Server 功能：
- 支持中间件、多种路由参数 HTTP Router ( xhttp.Router )
- HTTP Session 组件 ( store/xsession )
- 验证码功能：( ximage/caption )

通用基础库：
- 应用全局环境信息( xattr ): 管理应用基础环境信息如应用的根目录、配置文件目录、日志目录，数据目录等
- 支持多种格式的配置读取功能( xcfg )，支持从环境变量，应用全局环境信息( xattr )中读取配置值。
  1. 默认支持 .json 和  .xml
  2. .yaml  和  .toml 等其他格式可自行注册对应的驱动。 
- 支持泛型的缓存组件（store/xcache），已内置支持：
  1. File Storage：本地文件系统存储
  2. Memory Storage：内存存储，支持 LRU（最少使用先过期）、FIFO(先写入先过期)、LIFO(后写入先过期)
  3. Nop：黑洞，总是能成功写入但是读取不到值
  4. Redis Storage：使用 redis 作为缓存
  5. Chains：多层级缓存，可将多种 Cache 组合，设置不同的缓存容量和过期时间，以提升 Cache 效率
     - 如将 LRU Cache 和 Redis Cache 组合
  6. Reader：将数据源和 Cache 直接组合，透明的读取数据，而不需要先读取缓再设置缓存的模版代码
- 通用的，仿 Redis 存储的 Key-Value API，已支持数据类型 String、List、Hash、Set，ZSet。已内置支持驱动：
  1. File Storage：本地文件系统存储
  2. Memory Storage：内存存储
  3. Redis Storage：使用 redis 作为缓存
  4. Nop：黑洞，总是能成功写入但是读取不到值
  5. 当使用 K-V 存储时，使用此，可避免和 Redis 等具体的实现绑定，轻松配置管理实际存储方案。
- 日志库 （xlog）：基于标准库 slog 封装，轻量好用
- 支持泛型的轻量的单侧 assert 库 （ xt ）
- 国际化（i18n）的能力 （ xi18n ）

以上，每一个部分都能独立的使用