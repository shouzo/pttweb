package richcontent

import (
	"fmt"
	"html"
	"regexp"
	"strings"

	"golang.org/x/net/context"
)

func FindUrl(ctx context.Context, input []byte) ([]RichContent, error) {
	rcs := make([]RichContent, 0, 4)
	for _, u := range FindAllUrlsIndex(input) {
		urlBytes := input[u[0]:u[1]]
		var components []Component
		for _, p := range defaultUrlPatterns {
			if match := p.Pattern.FindSubmatchIndex(urlBytes); match != nil {
				if c, err := p.Handler(ctx, urlBytes, MatchIndices(match)); err == nil {
					components = c
				}
				break
			}
		}
		rcs = append(rcs, MakeRichContent(u[0], u[1], string(urlBytes), components))
	}
	return rcs, nil
}

type UrlPatternHandler func(ctx context.Context, urlBytes []byte, match MatchIndices) ([]Component, error)

type UrlPattern struct {
	Pattern *regexp.Regexp
	Handler UrlPatternHandler
}

var defaultUrlPatterns = []*UrlPattern{
	newUrlPattern(`^https?://(?:www\.youtube\.com/watch\?(?:.+&)*v=|youtu\.be/)([\w\-]+)`, handleYoutube),
	newUrlPattern(`^https?:(//i\.imgur\.com/[\.\w]+)$`, handleSameSchemeImage), // Note: cuz some users use http
	newUrlPattern(`^https?://imgur\.com/([,\w]+)(?:\#(\d+))?[^/]*$`, handleImgur),
	newUrlPattern(`^http://picmoe\.net/d\.php\?id=(\d+)`, handlePicmoe),
	newUrlPattern(`\.(?i:png|jpg|gif)$`, handleGenericImage),
}

func newUrlPattern(pattern string, handler UrlPatternHandler) *UrlPattern {
	return &UrlPattern{
		Pattern: regexp.MustCompile(pattern),
		Handler: handler,
	}
}

func imageHtmlTag(urlString string) string {
	return fmt.Sprintf(`<img src="%s" alt="" />`, html.EscapeString(urlString))
}

// Handlers

func handleYoutube(ctx context.Context, urlBytes []byte, match MatchIndices) ([]Component, error) {
	return []Component{MakeComponent(fmt.Sprintf(
		`<div class="resize-container"><div class="resize-content"><iframe class="youtube-player" type="text/html" src="//www.youtube.com/embed/%s" frameborder="0"></iframe></div></div>`,
		string(match.ByteSliceOf(urlBytes, 1))))}, nil
}

func handleSameSchemeImage(ctx context.Context, urlBytes []byte, match MatchIndices) ([]Component, error) {
	return []Component{MakeComponent(imageHtmlTag(string(match.ByteSliceOf(urlBytes, 1))))}, nil
}

func handleImgur(ctx context.Context, urlBytes []byte, match MatchIndices) ([]Component, error) {
	var comps []Component
	for _, id := range strings.Split(string(match.ByteSliceOf(urlBytes, 1)), ",") {
		link := fmt.Sprintf(`//i.imgur.com/%s.jpg`, id)
		comps = append(comps, MakeComponent(imageHtmlTag(link)))
	}
	return comps, nil
}

func handlePicmoe(ctx context.Context, urlBytes []byte, match MatchIndices) ([]Component, error) {
	link := fmt.Sprintf(`http://picmoe.net/src/%ss.jpg`, string(match.ByteSliceOf(urlBytes, 1)))
	return []Component{MakeComponent(imageHtmlTag(link))}, nil
}

func handleGenericImage(ctx context.Context, urlBytes []byte, match MatchIndices) ([]Component, error) {
	return []Component{MakeComponent(imageHtmlTag(string(urlBytes)))}, nil
}
