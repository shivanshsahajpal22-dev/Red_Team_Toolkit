package parsers

import (
	"strings"

	"golang.org/x/net/html"
)

type XParser struct{}

func NewXParser() *XParser {
	return &XParser{}
}

func (p *XParser) Name() string {
	return "x"
}

func (p *XParser) Parse(
	target Target,
	body []byte,
	statusCode int,
) (Profile, error) {

	profile := Profile{
		Platform:    "x",
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

	extractX(doc, &profile)

	if containsUsername(
		doc,
		target.Identifier,
	) {

		profile.Status = StatusExists
	}

	return profile, nil
}

func extractX(
	node *html.Node,
	profile *Profile,
) {

	if node.Type == html.ElementNode {

		switch strings.ToLower(node.Data) {

		case "title":

			if profile.DisplayName == "" {
				profile.DisplayName =
					nodeText(node)
			}

		case "a":

			for _, attr := range node.Attr {

				if attr.Key == "href" &&
					strings.HasPrefix(
						attr.Val,
						"http",
					) {

					profile.Links = append(
						profile.Links,
						attr.Val,
					)
				}
			}
		}
	}

	for child := node.FirstChild;
		child != nil;
		child = child.NextSibling {

		extractX(
			child,
			profile,
		)
	}
}

func containsUsername(
	node *html.Node,
	username string,
) bool {

	username = strings.ToLower(username)

	if node.Type == html.TextNode {

		if strings.Contains(
			strings.ToLower(node.Data),
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

func nodeText(node *html.Node) string {

	var result strings.Builder

	var walk func(*html.Node)

	walk = func(n *html.Node) {

		if n.Type == html.TextNode {
			result.WriteString(n.Data)
			result.WriteString(" ")
		}

		for child := n.FirstChild;
			child != nil;
			child = child.NextSibling {

			walk(child)
		}
	}

	walk(node)

	return strings.Join(
		strings.Fields(result.String()),
		" ",
	)
}
