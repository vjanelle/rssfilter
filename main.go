package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/beevik/etree"
)

const fallbackFeedURL = "https://boingboing.net/feed"
const fallbackBlockedCreators = "Boing Boing's Shop"

type FilterConfig struct {
	BlockedCreators   []string
	BlockedCategories []string
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRoot)
	mux.HandleFunc("/proxy/feed", handleProxyFeed)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	log.Printf("rss proxy listening on %s", addr)

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	cfg := configFromRequest(r)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = io.WriteString(w, "<h1>RSS Feed Proxy</h1>\n")
	_, _ = io.WriteString(w, "<p>Use <code>/proxy/feed</code> to get a filtered RSS feed.</p>\n")
	_, _ = io.WriteString(w, fmt.Sprintf("<p>Example: <a href=\"/proxy/feed?url=%s\">/proxy/feed?url=%s</a></p>\n", cfg.FeedURL, cfg.FeedURL))
	_, _ = io.WriteString(w, fmt.Sprintf("<p>Blocked creators: <strong>%s</strong></p>\n", strings.Join(cfg.Filter.BlockedCreators, ", ")))
	if len(cfg.Filter.BlockedCategories) > 0 {
		_, _ = io.WriteString(w, fmt.Sprintf("<p>Blocked categories: <strong>%s</strong></p>\n", strings.Join(cfg.Filter.BlockedCategories, ", ")))
	}
}

func handleProxyFeed(w http.ResponseWriter, r *http.Request) {
	cfg := configFromRequest(r)

	xmlBytes, status, err := fetch(cfg.FeedURL)
	if err != nil {
		log.Printf("fetch error url=%s: %v", cfg.FeedURL, err)
		http.Error(w, "Error fetching RSS feed", http.StatusBadGateway)
		return
	}
	if status < 200 || status >= 300 {
		log.Printf("fetch non-2xx url=%s status=%d", cfg.FeedURL, status)
		http.Error(w, "Upstream returned non-2xx", http.StatusBadGateway)
		return
	}

	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(xmlBytes); err != nil {
		log.Printf("xml parse error url=%s: %v", cfg.FeedURL, err)
		http.Error(w, "Error parsing RSS feed XML", http.StatusBadGateway)
		return
	}

	removed := filter(doc, cfg.Filter)

	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	w.Header().Set("X-Removed-Items", fmt.Sprintf("%d", removed))

	out, err := doc.WriteToBytes()
	if err != nil {
		log.Printf("xml serialize error url=%s: %v", cfg.FeedURL, err)
		http.Error(w, "Error serializing RSS feed XML", http.StatusInternalServerError)
		return
	}

	_, _ = w.Write(out)
}

func fetch(url string) ([]byte, int, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("User-Agent", "rssproxy/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return b, resp.StatusCode, nil
}

type RequestConfig struct {
	FeedURL string
	Filter  FilterConfig
}

func configFromRequest(r *http.Request) RequestConfig {
	feedURL := r.URL.Query().Get("url")
	if feedURL == "" {
		feedURL = strings.TrimSpace(os.Getenv("FEED_URL_DEFAULT"))
	}
	if feedURL == "" {
		feedURL = fallbackFeedURL
	}

	blockedCreators := queryOrEnvList(r, "blocked_creators", "BLOCKED_CREATORS", fallbackBlockedCreators)
	blockedCategories := queryOrEnvList(r, "blocked_categories", "BLOCKED_CATEGORIES", "")

	return RequestConfig{
		FeedURL: feedURL,
		Filter: FilterConfig{
			BlockedCreators:   blockedCreators,
			BlockedCategories: blockedCategories,
		},
	}
}

func queryOrEnvList(r *http.Request, queryKey string, envKey string, fallback string) []string {
	if raw := strings.TrimSpace(r.URL.Query().Get(queryKey)); raw != "" {
		return splitCSV(raw)
	}
	if raw := strings.TrimSpace(os.Getenv(envKey)); raw != "" {
		return splitCSV(raw)
	}
	if strings.TrimSpace(fallback) == "" {
		return nil
	}
	return splitCSV(fallback)
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}

func filter(doc *etree.Document, cfg FilterConfig) int {
	removed := 0
	items := doc.FindElements("//item")
	for _, item := range items {
		if !shouldRemoveItem(item, cfg) {
			continue
		}
		parent := item.Parent()
		if parent == nil {
			continue
		}
		parent.RemoveChild(item)
		removed++
	}
	return removed
}

func shouldRemoveItem(item *etree.Element, cfg FilterConfig) bool {
	if len(cfg.BlockedCreators) > 0 {
		creator := item.FindElement("dc:creator")
		if creator != nil {
			creatorText := strings.TrimSpace(creator.Text())
			for _, blocked := range cfg.BlockedCreators {
				if creatorText == blocked {
					return true
				}
			}
		}
	}

	if len(cfg.BlockedCategories) > 0 {
		cats := item.FindElements("category")
		for _, c := range cats {
			catText := strings.TrimSpace(c.Text())
			for _, blocked := range cfg.BlockedCategories {
				if catText == blocked {
					return true
				}
			}
		}
	}

	return false
}
