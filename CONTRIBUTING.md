# 贡献指南

感谢你对 Codeforces Go Client 的关注！

## 开发环境

```bash
git clone https://github.com/laoin114514/codeforcesClient.git
cd codeforcesClient
go mod download
```

## 测试

```bash
# 单元测试
go test ./...

# 集成测试（需要联网）
go test -v -count=1 ./test/

# 运行完整 API 验证
go run ./test/
```

## 代码风格

- 函数、类型、方法使用**中文注释**
- 错误处理使用 `errors.As` + `CFError.Code` 区分类型
- API 方法按分类组织：Blog → Contest → ProblemSet → User → 其他

## 添加新 API

1. 在 `models.go` 中定义 `XxxParams` 和 `XxxResponse` 类型
2. 在 `client.go` 对应分类下添加方法，调用 `c.doRequest(c.ctx, "method.name", params, nil, &resp)`
3. 在 `test/api_test.go` 中添加集成测试
4. 在 `docs/api.md` 中补文档

## 提交规范

```
feat: 新增功能
fix: 修复 bug
refactor: 重构
docs: 文档
test: 测试
chore: 杂项
```

## 开源协议

MIT License — 详见 [LICENSE](LICENSE)
