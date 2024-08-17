# ohbubblechat

Golang 实现的 IM 即时通讯服务，采用微服务架构，各服务之间基于 etcd 服务发现。

:strawberry: TODO：有空用 bubbletea 写一个简单的 terminal 玩具～

## 有哪些功能？
- 登陆注册、身份验证；
- 聊天室广播消息；
- 加入房间后，房间内成员私信消息；
- 获取聊天的房间信息：获取房间内在线人数，获取房间内用户的基本信息；
- 支持 http 连接和 websocket 连接。

## 使用了哪些技术栈？
- web 框架：[gin](https://github.com/gin-gonic/gin)；
- orm 框架：[gorm](https://github.com/go-gorm/gorm)；
- 微服务组件：rpc 框架 [rpcx](https://github.com/smallnest/rpcx)、[etcd](https://github.com/rpcxio/rpcx-etcd) 服务发现；
- 数据存储：mysql、redis；
- 消息队列：基于 redis list 数据结构；

## 建议关注的实现细节？
1. 微服务拆分：不同类型的连接之间的消息互通及消息的可靠推送。Redis 作为消息代理处理消息的存储和交付，上层使用 Gin 框架处理请求，消息则投放到消息队列，再由 rpc 广播至最上层连接层，最后由连接层将消息投递到对应的远端用户。
2. 会话连接管理：由唯一 id 标识房间和客户端会话连接，房间内的客户端连接使用链表管理，服务端定时发送下行心跳包检测客户端存活性。
3. goroutine 管理：如何保证 goroutine 被合理创建，并且不会发生 goroutine 泄漏？ 
4. 并发场景的应对：在连接层划分 bucket 减少高并发时的锁竞争。


