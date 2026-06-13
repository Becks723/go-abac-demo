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

## API 文档
https://s.apifox.cn/94624555-e8fb-44f7-a53c-14bdd7f081da

## 规则扩展

访问控制链路是：

`HTTP handler -> Service.CheckAccess -> Enforcer.Enforce -> Rule`

新增规则时实现 `Rule` 接口，然后注册到 `Enforcer` 即可，不需要改 `CheckAccess`：

```go
rule := abac.NewRuleFunc("vip-user", func(ctx abac.AccessContext) abac.RuleResult {
    if ctx.User.ID == "vip-user" {
        return abac.RuleResult{
            Effect: abac.EffectAllow,
            Status: abac.StatusAllowed,
            Reason: "vip user allowed",
        }
    }
    return abac.RuleResult{Effect: abac.EffectAbstain}
})

_ = svc.Enforcer().AddRule(rule)
svc.Enforcer().RemoveRule("region")
```
