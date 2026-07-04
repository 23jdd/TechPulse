# TechPulse Web 应用

gateway 启动后打开用户应用：

```text
http://localhost:8080/
http://localhost:8080/app
```

登录页：

```text
http://localhost:8080/login
```

中文页面：

```text
http://localhost:8080/login/zh
http://localhost:8080/zh
http://localhost:8080/app/zh
```

这个应用由 gateway 提供，是基于 Tailwind 的用户界面。它支持核心阅读流程：

- 添加 RSS Feed
- 测试、启用、停用、删除 Feed
- 抓取 RSS 和 GitHub Release
- 阅读文章信息流
- 搜索文章
- 查看 AI 摘要
- 标记已读、收藏、稍后读、归档、删除文章
- 使用带引用来源的 RAG 问答
- 生成每日技术报告
