# 搜索设计

TechPulse 的搜索设计保持简单、可解释、本地优先。

## 目标

- 支持搜索文章标题、正文、摘要和标签
- 返回高亮片段
- 支持分页和过滤
- 保持本地 demo 易运行
- 为未来向量搜索预留空间

## 当前实现

1. 关键词搜索：使用 Bleve 全文搜索
2. 字段权重：标题权重大于摘要、正文和标签
3. 过滤：支持标签、作者、来源、日期等过滤
4. 高亮：返回 Bleve 匹配片段
5. 分页：`page` 和 `page_size`
6. 可选重排：`HybridEngine` 可以调用 AI Provider 的 embedding 函数，对关键词召回结果重排
7. 可解释性：`/api/v1/search/explain` 返回当前字段、过滤条件和排序策略
8. 运维接口：`/api/v1/search/reindex` 从 MySQL 重建 Bleve 索引

## 为什么选择 Bleve

Bleve 是嵌入式、Go 原生搜索引擎。它避免 MVP 阶段引入 Elasticsearch，同时仍能展示搜索架构和 API 设计能力。

## 为什么做 Hybrid Search

关键词搜索精确且可解释，但容易漏掉语义相近的内容。Embedding 重排可以逐步引入语义检索，同时保留 Bleve 作为稳定召回层。

当前流程：

```text
query -> Bleve recall -> top hits -> embedding cosine rerank -> response
```

## 后续增强

- 更详细的 BM25 debug 解释
- 持久化向量索引
- 基于标签和摘要做 query expansion
- 用户级搜索个性化
