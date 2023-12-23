# Terminal Chat

纯 Golang 实现的 IM 即时通讯的前后端。后端服务使用分层架构，各层服务单独部署，各层之间基于 Etcd 服务发现，使用 Docker 打包部署上线；前端使用 bubble TUI 库实现了一个轻量、美观的 terminal 应用提供给用户使用。

## :sparkling_heart: 技术栈
- web 框架 [gin](https://github.com/gin-gonic/gin)
- rpc 框架 [rpcx](https://github.com/smallnest/rpcx)
- [etcd](https://github.com/rpcxio/rpcx-etcd) 服务发现
- 关系型数据库 mysql
- 缓存数据库 redis
- 消息队列
- Docker 快速部署
- [bubbletea](https://github.com/charmbracelet/bubbletea) TUI 工具库

## :cherry_blossom: 支持的连接
- HTTP
- Websocket

## :sunflower: 实现的功能
- 登陆注册、身份验证
- 聊天室广播消息
- 房间内成员私信消息
- 获取房间内在线人数
- 获取房间内用户的基本信息

## :blossom: 希望被关注的项目难点
1. 不同类型的连接之间的消息互通及消息的可靠推送 —— Redis 作为消息代理处理消息的存储和交付，上层使用 Gin 框架处理请求，消息则投放到消息队列，再由 rpc 广播至最上层连接层，最后由连接层将消息投递到对应的远端用户。
2. 客户端会话连接管理一 —— 由唯一id 标识房间和客户端会话连接，房间内的客户端连接使用链表管理，服务端定时发送下行心跳包检测客户端存活性。
3. 并发场景的应对 —— 在连接层划分 bucket 减少高并发时的锁竞争。


