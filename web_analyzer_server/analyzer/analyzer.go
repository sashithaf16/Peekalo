package analyzer

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/sashithaf16/peekalo/config"
	"github.com/sashithaf16/peekalo/logger"
	"golang.org/x/net/html"
)

type HttpClientInterface interface {
	Do(req *http.Request) (*http.Response, error)
}

type Analyzer struct {
	logger     logger.Logger
	cfg        *config.Config
	httpClient HttpClientInterface
}

type PageInfo struct {
	HTMLVersion string         `json:"html_version"`
	Title       string         `json:"title"`
	Headings    map[string]int `json:"headings"`
	Links       LinkStats      `json:"link_stats"`
	HasLogin    bool           `json:"has_login"`
}

type LinkStats struct {
	Internal     int `json:"internal"`
	External     int `json:"external"`
	Inaccessible int `json:"inaccessible"`
}

func NewAnalyzer(logger logger.Logger, cfg *config.Config, httpClient HttpClientInterface) *Analyzer {
	return &Analyzer{logger: logger, cfg: cfg, httpClient: httpClient}
}

func (a *Analyzer) AnalyzeURL(ctx context.Context, pageURL string) (PageInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", pageURL, nil)
	if err != nil {
		return PageInfo{}, fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := a.httpClient.Do(req)

	if err != nil {
		a.logger.Error().Err(err).Msgf("failed to fetch URL: %s", pageURL)
		return PageInfo{}, fmt.Errorf("failed to fetch URL: %v", err)
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		a.logger.Error().Err(err).Msgf("failed to parse HTML for URL: %s", pageURL)
		return PageInfo{}, fmt.Errorf("failed to parse HTML: %v", err)
	}
	parsedURL, err := url.Parse(pageURL)
	if err != nil {
		a.logger.Error().Err(err).Msgf("Invalid URL: %s", pageURL)
		return PageInfo{}, fmt.Errorf("invalid base URL: %v", err)
	}

	var wg sync.WaitGroup
	versionCh := make(chan string, 1)
	titleCh := make(chan string, 1)
	headingCh := make(chan map[string]int, 1)
	linksCh := make(chan LinkStats, 1)
	loginCh := make(chan bool, 1)

	wg.Add(5)
	go a.getHTMLVersion(ctx, doc, versionCh, &wg)
	go a.getPageTitle(ctx, doc, titleCh, &wg)
	go a.getHeadingsCount(ctx, doc, headingCh, &wg)
	go a.getLinkStats(ctx, doc, linksCh, &wg, parsedURL)
	go a.detectLoginForm(ctx, doc, loginCh, &wg)
	wg.Wait()
	a.logger.Debug().Msg("All analysis goroutines completed")

	info := PageInfo{
		HTMLVersion: <-versionCh,
		Title:       <-titleCh,
		Headings:    <-headingCh,
		Links:       <-linksCh,
		HasLogin:    <-loginCh,
	}
	return info, nil
}

func (a *Analyzer) getHeadingsCount(ctx context.Context, doc *html.Node, ch chan<- map[string]int, wg *sync.WaitGroup) {
	a.logger.Debug().Msg("Analyzing headings count")
	defer wg.Done()

	if isCancelled(ctx) {
		return
	}

	headings := map[string]int{
		"h1": 0, "h2": 0, "h3": 0, "h4": 0, "h5": 0, "h6": 0,
	}
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if _, exists := headings[n.Data]; exists {
				headings[n.Data]++
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)
	// Only send if context is still active

	if isCancelled(ctx) {
		return
	}
	ch <- headings
}

func (a *Analyzer) getPageTitle(ctx context.Context, doc *html.Node, ch chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	if isCancelled(ctx) {
		return
	}

	// Find the <head> node
	var head *html.Node
	for c := doc.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "html" {
			for hc := c.FirstChild; hc != nil; hc = hc.NextSibling {
				if hc.Type == html.ElementNode && hc.Data == "head" {
					head = hc
					break
				}
			}
			break
		}
	}

	var title string
	if head != nil {
		for c := head.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && c.Data == "title" && c.FirstChild != nil {
				title = c.FirstChild.Data
				break
			}
		}
	}

	if isCancelled(ctx) {
		return
	}
	ch <- title
}

func (a *Analyzer) getHTMLVersion(ctx context.Context, doc *html.Node, ch chan<- string, wg *sync.WaitGroup) {
	a.logger.Debug().Msg("Analyzing HTML version")
	defer wg.Done()

	if isCancelled(ctx) {
		return
	}

	version := "Unknown"

	for c := doc.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.DoctypeNode {
			doctype := strings.ToLower(c.Data)
			switch {
			case strings.Contains(doctype, "html 2.0"):
				version = "HTML 2"
			case strings.Contains(doctype, "html 3.2"):
				version = "HTML 3"
			case strings.Contains(doctype, "html 4.01"):
				version = "HTML 4"
			case strings.Contains(doctype, "html"):
				version = "HTML 5"
			}
			break
		}
	}

	if isCancelled(ctx) {
		return
	}
	ch <- version

}

func (a *Analyzer) getLinkStats(ctx context.Context, doc *html.Node, ch chan<- LinkStats, wg *sync.WaitGroup, baseURL *url.URL) {
	a.logger.Debug().Msg("Analyzing link statistics")
	defer wg.Done()

	if isCancelled(ctx) {
		return
	}

	var stats LinkStats

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key != "href" {
					continue
				}

				href := strings.TrimSpace(attr.Val)
				if href == "" || strings.HasPrefix(href, "#") {
					stats.Inaccessible++
					break
				}

				linkURL, err := url.Parse(href)
				if err != nil {
					stats.Inaccessible++
					break
				}

				resolved := baseURL.ResolveReference(linkURL)

				scheme := strings.ToLower(resolved.Scheme)
				switch scheme {
				case "http", "https":
					if strings.EqualFold(resolved.Host, baseURL.Host) {
						stats.Internal++
					} else {
						stats.External++
					}
				default:
					stats.Inaccessible++
				}

				break
			}
		}

		for child := n.FirstChild; child != nil; child = child.NextSibling {
			traverse(child)
		}
	}

	traverse(doc)
	if isCancelled(ctx) {
		return
	}
	ch <- stats
}

// logic - Looks for a <form> element with either a password input or a link containing "login" text
// html reference - https://www.w3schools.com/howto/howto_css_social_login.asp
func (a *Analyzer) detectLoginForm(ctx context.Context, doc *html.Node, ch chan<- bool, wg *sync.WaitGroup) {
	defer wg.Done()

	if isCancelled(ctx) {
		return
	}

	var found bool

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if found || n == nil {
			return
		}

		if n.Type == html.ElementNode && n.Data == "form" {
			if containsPasswordInput(n) || containsLoginAnchor(n) {
				found = true
				return
			}
		}

		for c := n.FirstChild; c != nil && !found; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)

	if isCancelled(ctx) {
		return
	}
	ch <- found
}

func containsPasswordInput(n *html.Node) bool {
	var found bool
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if found || node == nil {
			return
		}
		if node.Type == html.ElementNode && node.Data == "input" {
			for _, attr := range node.Attr {
				if attr.Key == "type" && strings.EqualFold(attr.Val, "password") {
					found = true
					return
				}
			}
		}
		for c := node.FirstChild; c != nil && !found; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return found
}

func containsLoginAnchor(n *html.Node) bool {
	var found bool
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if found || node == nil {
			return
		}
		if node.Type == html.ElementNode && node.Data == "a" {
			if strings.Contains(strings.ToLower(getText(node)), "login") || strings.Contains(strings.ToLower(getText(node)), "sign in") || strings.Contains(strings.ToLower(getText(node)), "continue with") {
				found = true
				return
			}
		}
		for c := node.FirstChild; c != nil && !found; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return found
}

func getText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		sb.WriteString(getText(c))
	}
	return sb.String()
}

func isCancelled(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
