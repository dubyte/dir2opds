package service

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"path"
	"strings"
)

const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Catalog.Title}} - dir2opds</title>
    <style>
        :root {
            --primary-color: #2c3e50;
            --secondary-color: #34495e;
            --accent-color: #3498db;
            --text-color: #333;
            --bg-color: #f4f7f6;
            --card-bg: #fff;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
            background-color: var(--bg-color);
            color: var(--text-color);
            line-height: 1.6;
            margin: 0;
            padding: 0;
        }
        .container {
            max-width: 900px;
            margin: 0 auto;
            padding: 20px;
        }
        header {
            background-color: var(--primary-color);
            color: white;
            padding: 20px 0;
            margin-bottom: 30px;
            box-shadow: 0 2px 5px rgba(0,0,0,0.1);
        }
        header h1 {
            margin: 0;
            text-align: center;
            font-size: 1.5rem;
        }
        .breadcrumb {
            margin-bottom: 20px;
            font-size: 0.9rem;
        }
        .breadcrumb a {
            color: var(--accent-color);
            text-decoration: none;
        }
        .breadcrumb span {
            margin: 0 5px;
            color: #999;
        }
        .search-box {
            margin-bottom: 30px;
            text-align: center;
        }
        .search-box input[type="text"] {
            padding: 10px;
            width: 60%;
            border: 1px solid #ddd;
            border-radius: 4px 0 0 4px;
            outline: none;
        }
        .search-box button {
            padding: 10px 20px;
            background-color: var(--accent-color);
            color: white;
            border: none;
            border-radius: 0 4px 4px 0;
            cursor: pointer;
        }
        .entry-list {
            list-style: none;
            padding: 0;
        }
        .entry-item {
            background-color: var(--card-bg);
            margin-bottom: 15px;
            padding: 15px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.05);
            display: flex;
            align-items: center;
            transition: transform 0.2s;
        }
        .entry-item:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 8px rgba(0,0,0,0.1);
        }
        .entry-icon {
            font-size: 2rem;
            margin-right: 20px;
            width: 50px;
            text-align: center;
            flex-shrink: 0;
        }
        .entry-cover {
            width: 60px;
            height: 80px;
            object-fit: cover;
            margin-right: 20px;
            border-radius: 4px;
            box-shadow: 0 1px 3px rgba(0,0,0,0.2);
        }
        .entry-details {
            flex-grow: 1;
        }
        .entry-title {
            font-size: 1.1rem;
            font-weight: bold;
            margin-bottom: 5px;
        }
        .entry-title a {
            color: var(--primary-color);
            text-decoration: none;
        }
        .entry-title a:hover {
            color: var(--accent-color);
        }
        .entry-meta {
            font-size: 0.85rem;
            color: #777;
        }
        .pagination {
            display: flex;
            justify-content: center;
            margin-top: 30px;
            gap: 10px;
        }
        .pagination a {
            padding: 8px 15px;
            background-color: var(--card-bg);
            color: var(--accent-color);
            text-decoration: none;
            border-radius: 4px;
            border: 1px solid #ddd;
        }
        .pagination span.current {
            padding: 8px 15px;
            background-color: var(--accent-color);
            color: white;
            border-radius: 4px;
        }
        .footer {
            margin-top: 50px;
            text-align: center;
            font-size: 0.8rem;
            color: #999;
            padding-bottom: 20px;
        }
    </style>
</head>
<body>
    <header>
        <div class="container">
            <h1>dir2opds</h1>
        </div>
    </header>

    <div class="container">
        {{if .EnableSearch}}
        <div class="search-box">
            <form action="/search" method="get">
                <input type="text" name="q" placeholder="Search books..." value="{{.Query}}">
                <button type="submit">Search</button>
            </form>
        </div>
        {{end}}

        <div class="breadcrumb">
            <a href="/">Home</a>
            {{range .Breadcrumbs}}
                <span>/</span>
                <a href="{{.Path}}">{{.Name}}</a>
            {{end}}
        </div>

        <ul class="entry-list">
            {{range .Entries}}
            <li class="entry-item">
                {{if .CoverURL}}
                <img src="{{.CoverURL}}" class="entry-cover" alt="Cover">
                {{else}}
                <div class="entry-icon">
                    {{if eq .Type 0}}📄{{else}}📁{{end}}
                </div>
                {{end}}
                <div class="entry-details">
                    <div class="entry-title">
                        <a href="{{.Href}}">{{if .Title}}{{.Title}}{{else}}{{.Name}}{{end}}</a>
                    </div>
                    <div class="entry-meta">
                        {{if .Author}}By {{.Author}} | {{end}}
                        {{if .Size}}{{.SizeDisplay}} | {{end}}
                        Modified: {{.ModTimeDisplay}}
                    </div>
                </div>
            </li>
            {{else}}
            <p>No entries found.</p>
            {{end}}
        </ul>

        {{if gt .TotalPages 1}}
        <div class="pagination">
            {{if gt .CurrentPage 1}}
            <a href="{{.PrevPageURL}}">&laquo; Previous</a>
            {{end}}
            <span class="current">Page {{.CurrentPage}} of {{.TotalPages}}</span>
            {{if lt .CurrentPage .TotalPages}}
            <a href="{{.NextPageURL}}">Next &raquo;</a>
            {{end}}
        </div>
        {{end}}

        <div class="footer">
            Generated by <a href="https://github.com/dubyte/dir2opds" style="color: #999;">dir2opds</a>
        </div>
    </div>
</body>
</html>
`

type Breadcrumb struct {
	Name string
	Path string
}

type HTMLEntry struct {
	CatalogEntry
	Href           string
	CoverURL       string
	SizeDisplay    string
	ModTimeDisplay string
}

type HTMLData struct {
	Catalog      *Catalog
	Entries      []HTMLEntry
	Breadcrumbs  []Breadcrumb
	EnableSearch bool
	Query        string
	CurrentPage  int
	TotalPages   int
	PrevPageURL  string
	NextPageURL  string
}

func (s OPDS) renderHTML(w http.ResponseWriter, req *http.Request, catalog *Catalog) error {
	tmpl, err := template.New("catalog").Parse(htmlTemplate)
	if err != nil {
		return err
	}

	data := HTMLData{
		Catalog:      catalog,
		EnableSearch: s.EnableSearch,
		Query:        req.URL.Query().Get("q"),
		CurrentPage:  catalog.Page,
		TotalPages:   (catalog.Total + catalog.PageSize - 1) / catalog.PageSize,
	}

	// Breadcrumbs
	urlPath := strings.Trim(req.URL.Path, "/")
	if urlPath != "" {
		parts := strings.Split(urlPath, "/")
		current := ""
		for _, part := range parts {
			current += "/" + part
			data.Breadcrumbs = append(data.Breadcrumbs, Breadcrumb{
				Name: part,
				Path: current,
			})
		}
	}

	// Entries
	for _, entry := range catalog.Entries {
		var entryPath string
		if strings.HasPrefix(catalog.ID, "search:") {
			entryPath = "/" + entry.Name
		} else {
			entryPath = path.Join(req.URL.Path, entry.Name)
		}

		href := (&url.URL{Path: entryPath}).String()
		
		var coverURL string
		if s.ExtractMetadata && entry.CoverPath != "" && entry.Type == pathTypeFile {
			coverURL = "/cover?file=" + url.QueryEscape(entryPath)
		}

		data.Entries = append(data.Entries, HTMLEntry{
			CatalogEntry:   entry,
			Href:           href,
			CoverURL:       coverURL,
			SizeDisplay:    formatSize(entry.Size),
			ModTimeDisplay: entry.ModTime.Format("2006-01-02"),
		})
	}

	// Pagination
	if data.CurrentPage > 1 {
		data.PrevPageURL = buildPageURL(req.URL.Path, req.URL.Query(), data.CurrentPage-1)
	}
	if data.CurrentPage < data.TotalPages {
		data.NextPageURL = buildPageURL(req.URL.Path, req.URL.Query(), data.CurrentPage+1)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return tmpl.Execute(w, data)
}

func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
