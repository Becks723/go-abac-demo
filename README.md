# go-abac-demo

一个基于 ABAC 模型的动态权限 demo，使用 Hertz 提供 HTTP API。

## 运行

```bash
go mod tidy
go run ./cmd/server
```

默认监听 `:8888`。

## API 示例

权限检查：

```bash
curl -X POST http://localhost:8888/access/check \
  -H 'content-type: application/json' \
  -d '{"user_id":"u1","document_id":"doc1","action":"edit","region":"CN","hour":10}'
```

记录 UI 行为并更新积分：

```bash
curl -X POST http://localhost:8888/events \
  -H 'content-type: application/json' \
  -d '{"user_id":"u1","document_id":"doc1","action":"publish"}'
```

查询用户积分：

```bash
curl http://localhost:8888/users/u1/points
```
