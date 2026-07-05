const API = window.location.origin || "http://localhost:8080";
const TEXT = window.TECHPULSE_I18N || {};
const state = { articles: [], selectedID: 0 };

function t(key) {
  return TEXT[key] || key;
}

function $(id) {
  return document.getElementById(id);
}

function escapeHTML(value) {
  return String(value || "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}

function injectStyles() {
  const style = document.createElement("style");
  style.textContent = `
    .field{height:2.5rem;width:100%;border-radius:.375rem;border:1px solid #d9dee7;background:#fff;padding:0 .75rem;font-size:.875rem;outline:none}
    .field:focus{border-color:#0f766e;box-shadow:0 0 0 2px rgba(15,118,110,.10)}
    .nav-btn{display:inline-flex;height:2.5rem;align-items:center;justify-content:center;gap:.5rem;border-radius:.375rem;border:1px solid #d9dee7;background:#fff;padding:0 .75rem;font-size:.875rem;font-weight:500}
    .nav-btn:hover,.chip:hover,.icon-btn:hover{background:#f8fafc}
    .chip{display:inline-flex;height:2rem;align-items:center;gap:.35rem;border-radius:.375rem;border:1px solid #d9dee7;background:#fff;padding:0 .65rem;font-size:.75rem;color:#475467}
    .action-primary{display:inline-flex;height:2.5rem;align-items:center;justify-content:center;gap:.5rem;border-radius:.375rem;background:#0f766e;padding:0 .75rem;font-size:.875rem;font-weight:600;color:#fff}
    .action-primary:hover{background:#115e59}
    .icon-btn{display:inline-flex;height:2.5rem;width:2.5rem;align-items:center;justify-content:center;border-radius:.375rem}
    .article-row{display:block;width:100%;padding:1rem;text-align:left}
    .article-row:hover{background:#f8fafc}
    .article-row.active{background:#ecfdf5}
  `;
  document.head.appendChild(style);
}

async function request(path, options = {}) {
  const headers = { "Content-Type": "application/json", ...(options.headers || {}) };
  const token = localStorage.getItem("techpulse_api_token");
  if (token) headers.Authorization = `Bearer ${token}`;
  const res = await fetch(`${API}${path}`, { ...options, headers });
  if (!res.ok) {
    let message = await res.text();
    try {
      const parsed = JSON.parse(message);
      message = parsed.error || parsed.message || message;
    } catch (_) {}
    throw new Error(message || `${t("failed")}: ${res.status}`);
  }
  if (res.status === 204) return {};
  return res.json();
}

function toast(message, error = false) {
  const el = $("toast");
  el.textContent = message;
  el.className = `fixed bottom-4 left-1/2 z-30 max-w-[calc(100vw-32px)] -translate-x-1/2 rounded-md px-4 py-2 text-sm text-white shadow-lg ${error ? "bg-rose-700" : "bg-ink"}`;
  setTimeout(() => el.classList.add("hidden"), 2600);
}

function dateText(value) {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  return date.toLocaleDateString(undefined, { month: "short", day: "numeric" });
}

function setList(title, meta) {
  $("listTitle").textContent = title;
  $("listMeta").textContent = meta;
}

function loadSession() {
  let session = null;
  try {
    session = JSON.parse(localStorage.getItem("techpulse_session") || "null");
  } catch (_) {}
  if (!session) {
    $("session").textContent = t("localDemo");
  } else if (session.mode === "github") {
    $("session").textContent = session.user?.username || t("githubUser");
  } else if (session.mode === "token") {
    $("session").textContent = t("tokenMode");
  } else {
    $("session").textContent = t("localDemo");
  }
}

async function checkHealth() {
  try {
    const data = await request("/health");
    $("health").textContent = data.redis ? `${t("ok")} · redis ${data.redis}` : t("ok");
    $("health").className = "rounded-full border border-emerald-200 bg-emerald-50 px-2 py-1 text-xs text-emerald-700";
  } catch (err) {
    $("health").textContent = t("error");
    $("health").className = "rounded-full border border-rose-200 bg-rose-50 px-2 py-1 text-xs text-rose-700";
  }
}

async function loadTasks() {
  const el = $("tasks");
  if (!el) return;
  try {
    const data = await request("/api/v1/tasks?page_size=6");
    const tasks = data.items || [];
    el.innerHTML = tasks.length ? tasks.map(renderTask).join("") : `<p class="text-sm text-muted">${escapeHTML(t("emptyTasks"))}</p>`;
  } catch (_) {
    el.innerHTML = `<p class="text-sm text-muted">${escapeHTML(t("emptyTasks"))}</p>`;
  }
}

function renderTask(task) {
  const statusClass = {
    success: "bg-emerald-50 text-emerald-700 border-emerald-200",
    running: "bg-sky-50 text-sky-700 border-sky-200",
    pending: "bg-amber-50 text-amber-700 border-amber-200",
    retrying: "bg-amber-50 text-amber-700 border-amber-200",
    failed: "bg-rose-50 text-rose-700 border-rose-200"
  }[task.status] || "bg-slate-50 text-muted border-line";
  return `
    <div class="rounded-md border border-line p-3">
      <div class="flex items-center justify-between gap-2">
        <span class="truncate text-sm font-medium">#${escapeHTML(task.id)} ${escapeHTML(task.type || "task")}</span>
        <span class="rounded-full border px-2 py-0.5 text-[11px] ${statusClass}">${escapeHTML(task.status || "")}</span>
      </div>
      ${task.error_message ? `<p class="mt-2 max-h-10 overflow-hidden text-xs text-rose-700">${escapeHTML(task.error_message)}</p>` : ""}
    </div>`;
}

function itemsFrom(data) {
  return data.items || data.hits || [];
}

function articleSummary(article) {
  return article.summary || article.clean_content || article.raw_content || "";
}

function renderArticles(items) {
  state.articles = items || [];
  const list = $("articleList");
  if (!state.articles.length) {
    list.innerHTML = `<div class="p-6 text-sm text-muted">${escapeHTML(t("emptyArticles"))}</div>`;
    return;
  }
  list.innerHTML = state.articles.map((article) => {
    const active = article.id === state.selectedID ? " active" : "";
    const tags = (article.tags || []).slice(0, 3).map((tag) => `<span class="rounded-full bg-slate-100 px-2 py-1 text-[11px] text-muted">${escapeHTML(tag)}</span>`).join("");
    return `
      <button class="article-row${active}" onclick="showArticle(${article.id})">
        <div class="flex items-start justify-between gap-3">
          <h3 class="min-w-0 text-sm font-semibold leading-6">${escapeHTML(article.title || "Untitled")}</h3>
          <span class="shrink-0 text-xs text-muted">${escapeHTML(dateText(article.published_at || article.created_at))}</span>
        </div>
        <p class="mt-1 max-h-11 overflow-hidden text-sm leading-5 text-muted">${escapeHTML(articleSummary(article))}</p>
        <div class="mt-3 flex flex-wrap items-center gap-2">
          ${tags}
          ${article.author ? `<span class="text-xs text-muted">${escapeHTML(article.author)}</span>` : ""}
          ${article.source_type ? `<span class="text-xs text-muted">${escapeHTML(article.source_type)}</span>` : ""}
        </div>
      </button>`;
  }).join("");
}

function pathWithPageSize(path) {
  return path.includes("?") ? `${path}&page_size=30` : `${path}?page_size=30`;
}

async function loadArticles(filter = "") {
  try {
    setList(t("streamTitle"), t("streamMeta"));
    $("articleList").innerHTML = `<div class="p-6 text-sm text-muted">${escapeHTML(t("loading"))}</div>`;
    const data = await request(pathWithPageSize(`/api/v1/articles${filter}`));
    renderArticles(itemsFrom(data));
  } catch (err) {
    toast(err.message, true);
  }
}

async function loadFavorites(type) {
  try {
    setList(type === "read_later" ? t("laterTitle") : t("savedTitle"), type === "read_later" ? t("laterMeta") : t("savedMeta"));
    $("articleList").innerHTML = `<div class="p-6 text-sm text-muted">${escapeHTML(t("loading"))}</div>`;
    const data = await request(pathWithPageSize(`/api/v1/favorites?type=${encodeURIComponent(type)}`));
    renderArticles(itemsFrom(data));
  } catch (err) {
    toast(err.message, true);
  }
}

async function loadHistory() {
  try {
    setList(t("historyTitle"), t("historyMeta"));
    $("articleList").innerHTML = `<div class="p-6 text-sm text-muted">${escapeHTML(t("loading"))}</div>`;
    const data = await request(pathWithPageSize("/api/v1/reading-history"));
    renderArticles(itemsFrom(data));
  } catch (err) {
    toast(err.message, true);
  }
}

async function searchArticles(event) {
  if (event) event.preventDefault();
  const query = $("query").value.trim();
  const tag = $("tagFilter").value.trim();
  const params = new URLSearchParams({ q: query, tag, page_size: "30" });
  try {
    setList(t("searchTitle"), t("searchMeta"));
    $("articleList").innerHTML = `<div class="p-6 text-sm text-muted">${escapeHTML(t("loading"))}</div>`;
    const data = await request(`/api/v1/search?${params.toString()}`);
    renderArticles(itemsFrom(data));
  } catch (err) {
    toast(err.message, true);
  }
}

async function showArticle(id) {
  state.selectedID = id;
  const article = state.articles.find((item) => item.id === id);
  renderArticles(state.articles);
  if (!article) return;
  $("selected").innerHTML = `
    <div class="space-y-3">
      <div>
        <h3 class="text-base font-semibold leading-6">${escapeHTML(article.title || "Untitled")}</h3>
        <p class="mt-1 text-xs text-muted">${escapeHTML([article.author, dateText(article.published_at || article.created_at), article.source_type].filter(Boolean).join(" · "))}</p>
      </div>
      <div class="flex flex-wrap gap-2">
        <a href="${escapeHTML(article.url || "#")}" target="_blank" rel="noreferrer" class="chip"><i data-lucide="external-link" class="h-3.5 w-3.5"></i>${escapeHTML(t("open"))}</a>
        <button onclick="markRead(${id})" class="chip"><i data-lucide="check" class="h-3.5 w-3.5"></i>${escapeHTML(t("read"))}</button>
        <button onclick="favoriteArticle(${id})" class="chip"><i data-lucide="star" class="h-3.5 w-3.5"></i>${escapeHTML(t("save"))}</button>
        <button onclick="readLater(${id})" class="chip"><i data-lucide="clock-3" class="h-3.5 w-3.5"></i>${escapeHTML(t("later"))}</button>
        <button onclick="archiveArticle(${id})" class="chip"><i data-lucide="archive" class="h-3.5 w-3.5"></i>${escapeHTML(t("archive"))}</button>
        <button onclick="deleteArticle(${id})" class="chip"><i data-lucide="trash-2" class="h-3.5 w-3.5"></i>${escapeHTML(t("delete"))}</button>
      </div>
    </div>`;
  $("summary").classList.remove("hidden");
  $("summary").innerHTML = `<p class="text-muted">${escapeHTML(t("loading"))}</p>`;
  if (window.lucide) window.lucide.createIcons();
  try {
    const summary = await request(`/api/v1/articles/${id}/summary`);
    const bullets = (summary.bullet_points || "").split(/\n+/).filter(Boolean).map((item) => `<li>${escapeHTML(item.replace(/^[-*]\s*/, ""))}</li>`).join("");
    $("summary").innerHTML = `
      <p class="font-medium">${escapeHTML(summary.one_sentence || summary.tldr || t("noSummary"))}</p>
      ${summary.short_summary ? `<p class="mt-3 text-muted">${escapeHTML(summary.short_summary)}</p>` : ""}
      ${bullets ? `<ul class="mt-3 list-disc space-y-1 pl-5 text-muted">${bullets}</ul>` : ""}
      ${summary.long_summary ? `<p class="mt-3 text-muted">${escapeHTML(summary.long_summary)}</p>` : ""}`;
  } catch (_) {
    $("summary").innerHTML = `<p class="text-muted">${escapeHTML(t("noSummary"))}</p>`;
  }
}

async function mutateArticle(id, path, message, reload = false) {
  try {
    await request(`/api/v1/articles/${id}/${path}`, { method: "POST" });
    toast(message);
    if (reload) await loadArticles();
  } catch (err) {
    toast(err.message, true);
  }
}

function markRead(id) {
  mutateArticle(id, "read", t("readDone"));
}

function favoriteArticle(id) {
  mutateArticle(id, "favorite", t("saved"));
}

function readLater(id) {
  mutateArticle(id, "read-later", t("queued"));
}

function archiveArticle(id) {
  mutateArticle(id, "archive", t("archived"), true);
}

async function deleteArticle(id) {
  if (!confirm(t("delete"))) return;
  try {
    await request(`/api/v1/articles/${id}`, { method: "DELETE" });
    toast(t("deleted"));
    $("selected").textContent = t("selectedHint");
    $("summary").classList.add("hidden");
    await loadArticles();
  } catch (err) {
    toast(err.message, true);
  }
}

async function loadFeeds() {
  try {
    const data = await request("/api/v1/rss");
    const feeds = data.items || [];
    $("feeds").innerHTML = feeds.length ? feeds.map(renderFeed).join("") : `<p class="text-sm text-muted">${escapeHTML(t("emptyFeeds"))}</p>`;
    if (window.lucide) window.lucide.createIcons();
  } catch (err) {
    toast(err.message, true);
  }
}

function renderFeed(feed) {
  const enabled = feed.status !== "disabled";
  return `
    <div class="rounded-md border border-line p-3">
      <div class="flex items-start justify-between gap-3">
        <div class="min-w-0">
          <p class="truncate text-sm font-medium">${escapeHTML(feed.title || feed.url)}</p>
          <p class="mt-1 truncate text-xs text-muted">${escapeHTML(feed.category || feed.url)}</p>
        </div>
        <span class="rounded-full bg-slate-100 px-2 py-1 text-[11px] text-muted">${escapeHTML(feed.status || "active")}</span>
      </div>
      <div class="mt-3 grid grid-cols-5 gap-1">
        <button onclick="fetchFeed(${feed.id})" class="icon-btn border border-line" title="${escapeHTML(t("fetch"))}" aria-label="${escapeHTML(t("fetch"))}"><i data-lucide="download" class="h-4 w-4"></i></button>
        <button onclick="testFeed(${feed.id})" class="icon-btn border border-line" title="${escapeHTML(t("test"))}" aria-label="${escapeHTML(t("test"))}"><i data-lucide="radar" class="h-4 w-4"></i></button>
        <button onclick="setFeedStatus(${feed.id}, ${enabled ? "'disable'" : "'enable'"})" class="icon-btn border border-line" title="${escapeHTML(enabled ? t("disable") : t("enable"))}" aria-label="${escapeHTML(enabled ? t("disable") : t("enable"))}"><i data-lucide="${enabled ? "pause" : "play"}" class="h-4 w-4"></i></button>
        <a href="${escapeHTML(feed.url)}" target="_blank" rel="noreferrer" class="icon-btn border border-line" title="${escapeHTML(t("open"))}" aria-label="${escapeHTML(t("open"))}"><i data-lucide="external-link" class="h-4 w-4"></i></a>
        <button onclick="deleteFeed(${feed.id})" class="icon-btn border border-line" title="${escapeHTML(t("delete"))}" aria-label="${escapeHTML(t("delete"))}"><i data-lucide="trash-2" class="h-4 w-4"></i></button>
      </div>
    </div>`;
}

async function addFeed(event) {
  event.preventDefault();
  const payload = {
    url: $("feedUrl").value.trim(),
    title: $("feedTitle").value.trim(),
    category: $("feedCategory").value.trim(),
    fetch_interval_minutes: Number($("feedInterval").value || 60)
  };
  try {
    await request("/api/v1/rss", { method: "POST", body: JSON.stringify(payload) });
    $("feedUrl").value = "";
    $("feedTitle").value = "";
    $("feedCategory").value = "";
    toast(t("sourceAdded"));
    await loadFeeds();
  } catch (err) {
    toast(err.message, true);
  }
}

async function fetchFeed(id) {
  try {
    const data = await request(`/api/v1/rss/${id}/fetch-async`, { method: "POST" });
    toast(`${t("fetchQueued")} #${data.task_id}`);
    await loadTasks();
    pollTask(data.task_id);
  } catch (err) {
    toast(err.message, true);
  }
}

async function pollTask(id) {
  if (!id) return;
  for (let attempt = 0; attempt < 20; attempt += 1) {
    await new Promise((resolve) => setTimeout(resolve, 3000));
    try {
      const task = await request(`/api/v1/tasks/${id}`);
      if (task.status === "success") {
        toast(t("fetched"));
        await loadArticles();
        await loadTasks();
        return;
      }
      if (task.status === "failed") {
        toast(task.error_message || t("failed"), true);
        await loadTasks();
        return;
      }
    } catch (_) {
      return;
    }
  }
}

async function testFeed(id) {
  try {
    const data = await request(`/api/v1/rss/${id}/test`, { method: "POST" });
    toast(`${t("test")}: ${data.fetched || 0}`);
  } catch (err) {
    toast(err.message, true);
  }
}

async function setFeedStatus(id, action) {
  try {
    await request(`/api/v1/rss/${id}/${action}`, { method: "POST" });
    toast(t("sourceUpdated"));
    await loadFeeds();
  } catch (err) {
    toast(err.message, true);
  }
}

async function deleteFeed(id) {
  if (!confirm(t("delete"))) return;
  try {
    await request(`/api/v1/rss/${id}`, { method: "DELETE" });
    toast(t("sourceDeleted"));
    await loadFeeds();
  } catch (err) {
    toast(err.message, true);
  }
}

async function fetchGitHubReleases() {
  const url = $("githubUrl").value.trim();
  try {
    await request("/api/v1/github/releases/fetch", { method: "POST", body: JSON.stringify({ url }) });
    $("githubUrl").value = "";
    toast(t("fetched"));
    await loadArticles();
  } catch (err) {
    toast(err.message, true);
  }
}

async function fetchHackerNews() {
  const feed = $("hnFeed").value;
  const limit = Number($("hnLimit").value || 20);
  try {
    await request("/api/v1/hackernews/fetch", { method: "POST", body: JSON.stringify({ feed, limit }) });
    toast(t("fetched"));
    await loadArticles();
  } catch (err) {
    toast(err.message, true);
  }
}

async function ask() {
  const question = $("question").value.trim();
  if (!question) {
    toast(t("askRequired"), true);
    return;
  }
  $("answer").textContent = t("loading");
  $("citations").innerHTML = "";
  try {
    const data = await request("/api/v1/chat", { method: "POST", body: JSON.stringify({ question }) });
    $("answer").textContent = data.answer || "";
    $("citations").innerHTML = (data.citations || []).map((citation) => `
      <a href="${escapeHTML(citation.url || "#")}" target="_blank" rel="noreferrer" class="block rounded-md border border-line p-3 text-sm hover:bg-slate-50">
        <span class="font-medium">${escapeHTML(citation.title || `#${citation.article_id}`)}</span>
        ${citation.snippet ? `<span class="mt-1 block text-xs text-muted">${escapeHTML(citation.snippet)}</span>` : ""}
      </a>`).join("");
  } catch (err) {
    $("answer").textContent = "";
    toast(err.message, true);
  }
}

async function generateReport() {
  const title = $("reportTitle").value.trim() || t("reportTitle");
  $("report").textContent = t("loading");
  try {
    const data = await request("/api/v1/daily-reports", { method: "POST", body: JSON.stringify({ title }) });
    const report = data.report || data;
    $("report").textContent = report.content || "";
    toast(t("generated"));
  } catch (err) {
    $("report").textContent = "";
    toast(err.message, true);
  }
}

document.addEventListener("DOMContentLoaded", () => {
  injectStyles();
  loadSession();
  checkHealth();
  loadTasks();
  loadFeeds();
  loadArticles();
  if (window.lucide) window.lucide.createIcons();
});
