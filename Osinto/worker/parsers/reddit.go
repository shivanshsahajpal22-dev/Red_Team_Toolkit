package parsers

import (
	"strings"

	"golang.org/x/net/html"
)

type RedditParser struct{}

func NewRedditParser() *RedditParser {
	return &RedditParser{}
}

func (p *RedditParser) Name() string {
	return "reddit"
}

func (p *RedditParser) Parse(
	target Target,
	body []byte,
	statusCode int,
) (Profile, error) {

	profile := Profile{
		Platform:    "reddit",
		Identifier:  target.Identifier,
		Status:      StatusUnknown,
		URL: strings.ReplaceAll(
			target.PlatformURL,
			"{username}",
			target.Identifier,
		),
		Links: []string{},
	}

	if statusCode == 404 || statusCode == 410 {
		profile.Status = StatusNotFound
		return profile, nil
	}

	if statusCode == 429 {
		profile.Status = StatusRateLimited
		return profile, nil
	}

	if statusCode != 200 {
		return profile, nil
	}

	doc, err := html.Parse(
		strings.NewReader(string(body)),
	)

	if err != nil {
		return profile, err
	}

	/*
		Do NOT use the page title alone.

		Reddit frequently returns generic pages
		with HTTP 200.
	*/

	if containsUsername(
		doc,
		target.Identifier,
	) {

		profile.Status = StatusExists
	}

	return profile, nil
}

func containsUsername(
	node *html.Node,
	username string,
) bool {

	username = strings.ToLower(username)

	if node.Type == html.TextNode {

		text := strings.ToLower(
			node.Data,
		)

		if strings.Contains(
			text,
			username,
		) {
			return true
		}
	}

	for _, attr := range node.Attr {

		if strings.Contains(
			strings.ToLower(attr.Val),
			username,
		) {
			return true
		}
	}

	for child := node.FirstChild;
		child != nil;
		child = child.NextSibling {

		if containsUsername(
			child,
			username,
		) {
			return true
		}
	}

	return false
}
