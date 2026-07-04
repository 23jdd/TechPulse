# Search Design

TechPulse search is intentionally simple, explainable, and local-first.

## Goals

- Search article title, body, summary, and tags.
- Return highlight snippets.
- Support pagination and filters.
- Keep local demo easy to run.
- Leave room for future vector search.

## Current Implementation

1. Lexical search: Bleve full-text search.
2. Field boost: title is boosted above summary/content/tags.
3. Filters: tag and author filters are supported.
4. Highlight: matched fragments are returned from Bleve.
5. Pagination: `page` and `page_size`.
6. Optional reranking: `HybridEngine` can call the AI provider embedding function and rerank top lexical hits.
7. Explainability: `/api/v1/search/explain` returns active fields, filters, and ranking strategy.
8. Operations: `/api/v1/search/reindex` rebuilds the Bleve index from MySQL.

## Why Bleve

Bleve is embedded and Go-native. It avoids requiring Elasticsearch for an MVP while still proving search architecture and API design.

## Why Hybrid Search

Lexical search is precise and explainable, but it misses semantic matches. Embedding reranking gives a path toward semantic retrieval while keeping Bleve as the reliable recall layer.

Current hybrid flow:

```text
query -> Bleve recall -> top hits -> embedding cosine rerank -> response
```

## Future Improvements

- date range filters in the API handler
- BM25 score explanation in debug mode
- persistent vector index
- query expansion using tags and summaries
- per-user search personalization
