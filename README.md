# IM 服务

一个面向会话/消息/通知的 IM 后端服务，提供 gRPC 接口、消息队列消费与推送分发能力。

## 业务架构

- 会话管理：创建会话、获取会话信息、会话成员/智能体管理（群聊加入/查询）。
- 消息能力：发送消息、按会话/用户拉取消息、命令消息队列化分发。
- 已读与未读：会话维度的读索引、未读数（badge）维护与同步。
- 通知系统：通知发送、聚合通知、通知列表与未读统计、已读清零。
- 推送分发：消费消息后构建推送包，通过 Push 服务下发到用户端。

## 技术架构

- 接入层：gRPC 服务暴露 IM API（`proto/` + `proto_gen/`）。
- 事件驱动（双 Topic 分工）：
    - 会话 Topic（`IM_CONV_TOPIC`）：负责消息落库与会话链构建；根据会话类型（单聊/群聊/AI）计算接收者集合，并将消息扇出到用户
      Topic。
    - 用户 Topic（`IM_USER_TOPIC`）：为每个用户构建个人视角链路（会话链/命令链），维护未读数与索引位置，生成推送包并调用 Push
      服务下发。
- 三条链路（拉取数据的核心索引）：
    - 会话链（Conversation Index）：按会话维度有序存储消息 ID，支持按会话分页拉取历史消息。
    - 用户会话链（User Conversation Index）：按用户维度记录其参与会话的游标与顺序，用于拉取最近会话列表及增量更新。
    - 用户命令链（User Command Index）：按用户维度记录命令类消息（如 MarkRead/更新类）的消费位置，保证控制类消息的有序下发。
- 存储层：MySQL（会话/成员等元数据）、Redis（缓存/分布式锁/集合）、Kvrocks（消息体与索引链，分段存储与过期策略）。

## Proto 生成命令

```bash
protoc -I . --go_out=. --go_opt=module=im --go-grpc_out=. --go-grpc_opt=module=im .\proto\common.proto .\proto\im.proto .\proto\action.proto .\proto\push.proto
```
