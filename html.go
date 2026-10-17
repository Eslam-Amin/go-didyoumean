package didyoumean

import (
	"html/template"
	"net/http"

	"github.com/eslam-amin/didyoumean/types"
)

var (
	cardFuncMap = template.FuncMap{
		"multiply": func(a float64, b float64) float64 {
			return a * b
		},
	}
	errorCardTemplate = template.Must(template.New("error_card").Funcs(cardFuncMap).Parse(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>404 Not Found &mdash; Did You Mean?</title>
  <style>
    :root {
      --bg: #0d1117;
      --card-bg: #161b22;
      --card-border: #30363d;
      --text-main: #f0f6fc;
      --text-muted: #8b949e;
      --text-dim: #6e7681;
      --accent-blue: #58a6ff;
      --accent-red: #f85149;
      --accent-green: #3fb950;
      --accent-amber: #d29922;
      --accent-purple: #bc8cff;
      --badge-bg: #21262d;
    }
    * {
      box-sizing: border-box;
      margin: 0;
      padding: 0;
    }
    body {
      background-color: var(--bg);
      color: var(--text-main);
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
      min-height: 100vh;
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      padding: 2rem 1rem;
    }
    .container {
      width: 100%;
      max-width: 720px;
      background: var(--card-bg);
      border: 1px solid var(--card-border);
      border-radius: 12px;
      box-shadow: 0 10px 30px rgba(0, 0, 0, 0.5);
      overflow: hidden;
    }
    .header {
      padding: 1.5rem;
      border-bottom: 1px solid var(--card-border);
      background: linear-gradient(180deg, rgba(248, 81, 73, 0.08) 0%, transparent 100%);
    }
    .status-badge {
      display: inline-block;
      background: rgba(248, 81, 73, 0.2);
      color: var(--accent-red);
      font-weight: 700;
      font-size: 0.8rem;
      padding: 0.25rem 0.6rem;
      border-radius: 9999px;
      text-transform: uppercase;
      letter-spacing: 0.05em;
      margin-bottom: 0.75rem;
      border: 1px solid rgba(248, 81, 73, 0.3);
    }
    h1 {
      font-size: 1.5rem;
      font-weight: 600;
      margin-bottom: 0.5rem;
      color: #fff;
    }
    .requested-route {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      margin-top: 0.75rem;
      padding: 0.6rem 0.85rem;
      background: rgba(0, 0, 0, 0.3);
      border-radius: 8px;
      border: 1px solid rgba(255, 255, 255, 0.06);
      font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
      font-size: 0.95rem;
    }
    .method-tag {
      font-weight: 700;
      font-size: 0.75rem;
      padding: 0.2rem 0.45rem;
      border-radius: 4px;
      text-transform: uppercase;
    }
    .method-GET { background: rgba(63, 185, 80, 0.2); color: var(--accent-green); border: 1px solid rgba(63, 185, 80, 0.4); }
    .method-POST { background: rgba(210, 153, 34, 0.2); color: var(--accent-amber); border: 1px solid rgba(210, 153, 34, 0.4); }
    .method-PUT { background: rgba(88, 166, 255, 0.2); color: var(--accent-blue); border: 1px solid rgba(88, 166, 255, 0.4); }
    .method-DELETE { background: rgba(248, 81, 73, 0.2); color: var(--accent-red); border: 1px solid rgba(248, 81, 73, 0.4); }
    .method-PATCH { background: rgba(188, 140, 255, 0.2); color: var(--accent-purple); border: 1px solid rgba(188, 140, 255, 0.4); }
    .section-title {
      font-size: 0.9rem;
      color: var(--text-muted);
      text-transform: uppercase;
      letter-spacing: 0.05em;
      margin: 1.25rem 1.5rem 0.75rem;
      font-weight: 600;
    }
    .suggestions-list {
      list-style: none;
      padding: 0 1.5rem 1.5rem;
      display: flex;
      flex-direction: column;
      gap: 0.75rem;
    }
    .suggestion-card {
      background: var(--badge-bg);
      border: 1px solid var(--card-border);
      border-radius: 8px;
      padding: 0.85rem 1rem;
      transition: border-color 0.15s ease, background 0.15s ease;
    }
    .suggestion-card:hover {
      border-color: var(--accent-blue);
      background: #262c36;
    }
    .suggestion-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 0.75rem;
      margin-bottom: 0.4rem;
    }
    .suggestion-route {
      display: flex;
      align-items: center;
      gap: 0.6rem;
      font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
      font-size: 0.95rem;
      color: #fff;
    }
    .score-badge {
      font-size: 0.75rem;
      font-weight: 600;
      padding: 0.2rem 0.5rem;
      border-radius: 12px;
      background: rgba(88, 166, 255, 0.15);
      color: var(--accent-blue);
      border: 1px solid rgba(88, 166, 255, 0.3);
      white-space: nowrap;
    }
    .mismatch-badge {
      display: inline-flex;
      align-items: center;
      gap: 0.25rem;
      font-size: 0.75rem;
      font-weight: 600;
      color: var(--accent-amber);
      background: rgba(210, 153, 34, 0.15);
      border: 1px solid rgba(210, 153, 34, 0.3);
      padding: 0.15rem 0.45rem;
      border-radius: 4px;
      margin-top: 0.35rem;
    }
    .suggestion-meta {
      font-size: 0.825rem;
      color: var(--text-muted);
      margin-top: 0.25rem;
      line-height: 1.4;
    }
    .footer {
      padding: 0.9rem 1.5rem;
      border-top: 1px solid var(--card-border);
      background: rgba(0, 0, 0, 0.15);
      display: flex;
      align-items: center;
      justify-content: space-between;
      font-size: 0.775rem;
      color: var(--text-dim);
    }
    .footer a {
      color: var(--accent-blue);
      text-decoration: none;
    }
    .footer a:hover {
      text-decoration: underline;
    }
  </style>
</head>
<body>
  <div class="container">
    <div class="header">
      <div class="status-badge">404 Not Found</div>
      <h1>Endpoint Not Found</h1>
      <p style="color: var(--text-muted); font-size: 0.9rem;">The requested URL did not match any registered route handlers.</p>
      <div class="requested-route">
        <span class="method-tag method-{{.Method}}">{{.Method}}</span>
        <span>{{.Path}}</span>
      </div>
    </div>

    <div class="section-title">Did you mean?</div>
    <ul class="suggestions-list">
      {{range .Suggestions}}
      <li class="suggestion-card">
        <div class="suggestion-header">
          <div class="suggestion-route">
            <span class="method-tag method-{{.Method}}">{{.Method}}</span>
            <span>{{.Path}}</span>
          </div>
          <span class="score-badge">{{printf "%.0f%% Match" (multiply .Score 100)}}</span>
        </div>
        {{if .IsMethodMismatch}}
        <div class="mismatch-badge">
          &#9888; Method Mismatch (Path is registered for {{.Method}})
        </div>
        {{end}}
        {{if .Summary}}
        <div class="suggestion-meta"><strong>Summary:</strong> {{.Summary}}</div>
        {{end}}
        {{if .Description}}
        <div class="suggestion-meta">{{.Description}}</div>
        {{end}}
      </li>
      {{else}}
      <li style="color: var(--text-muted); padding: 0.5rem 0;">No close routes found.</li>
      {{end}}
    </ul>

    <div class="footer">
      <span>Powered by <strong>didyoumean</strong> (Go)</span>
      <span>Development Mode</span>
    </div>
  </div>
</body>
</html>`))
)

type errorCardData struct {
	Method      string
	Path        string
	Suggestions []types.RouteSuggestion
}

// RenderHTMLCard writes a developer-friendly HTML error card to the response writer.
func RenderHTMLCard(w http.ResponseWriter, r *http.Request, suggestions []types.RouteSuggestion) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)

	data := errorCardData{
		Method:      r.Method,
		Path:        r.URL.Path,
		Suggestions: suggestions,
	}

	return errorCardTemplate.Execute(w, data)
}
