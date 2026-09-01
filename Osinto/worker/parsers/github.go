package parsers

import (
	"strings"

	"osinto/worker/model"

	"golang.org/x/net/html"
)

type GitHubParser struct{}

func NewGitHubParser() *GitHubParser {
	return &GitHubParser{}
}

func (p *GitHubParser) Name() string {
	return "github"
}

func (p *GitHubParser) Parse(
	target model.Target,
	body []byte,
	statusCode int,
) (model.Profile, error) {

	profile := model.Profile{
		Platform:    "github",
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
		profile.Status = model.StatusNotFound
		return profile, nil
	}

	if statusCode == 429 {
		profile.Status = model.StatusRateLimited
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

	extractGitHub(doc, &profile)

	/*
		A generic HTTP 200 is NOT enough to declare
		the account valid.

		We require GitHub-specific evidence.
	*/

	if profile.DisplayName != "" ||
		len(profile.Links) > 0 {

		profile.Status = model.StatusExists
	}

	return profile, nil
}

func extractGitHub(
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

				if attr.Key != "href" {
					continue
				}

				if strings.HasPrefix(
					attr.Val,
					"https://github.com/",
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

		extractGitHub(
			child,
			profile,
		)
	}
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
