# TechPulse 请求说明（UTF-8 中文版）

原始 `Request.md` 是中文内容，但当前文件编码显示异常。此文件提供可读的 UTF-8 中文版本。

## 项目简介

TechPulse 是一个面向开发者的 AI 智能信息平台。系统持续收集来自 RSS、GitHub、HackerNews、Reddit、Arxiv、YouTube 等多个来源的技术内容，并利用 AI 自动完成：

- 内容解析
- 去重
- 标签分类
- 多语言翻译
- 摘要生成
- 全文索引
- 智能搜索
- RAG 问答
- 每日技术日报

最终目标是形成属于开发者自己的技术知识库。

## 目标

实现一个可以长期运行的开发者信息平台。它不只是 RSS 阅读器，而是 AI Knowledge Hub。

支持统一收集：

- 技术新闻
- GitHub Release
- AI 新闻
- Go 社区
- Linux
- Kubernetes
- 数据库
- 网络编程
- 云原生
- 安全

## 当前项目亮点

- Go 后端工程结构
- RSS / GitHub Releases 多源采集
- MySQL、Redis、RabbitMQ、etcd、MinIO 基础设施
- Bleve 全文搜索
- AI 摘要、标签、关键词、embedding
- RAG 问答并返回引用来源
- WebSocket 事件
- GitHub OAuth
- SMTP 日报邮件
- Tailwind Dashboard
- Docker Compose 一键运行

## 演示入口

```text
http://localhost:8080/
http://localhost:8080/zh
```
