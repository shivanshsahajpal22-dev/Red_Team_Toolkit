package main

import (
	"fmt"
	"strings"

	"osinto/worker/model"
	"osinto/worker/parsers"
)

type Parser interface {
	Name() string

	Parse(
		target model.Target,
		body []byte,
		statusCode int,
	) (model.Profile, error)
}

type ParserRegistry struct {
	parsers map[string]Parser
}

func NewParserRegistry() *ParserRegistry {
	return &ParserRegistry{
		parsers: make(map[string]Parser),
	}
}

func (r *ParserRegistry) Register(parser Parser) {

	name := strings.ToLower(
		strings.TrimSpace(parser.Name()),
	)

	r.parsers[name] = parser
}

func (r *ParserRegistry) Get(
	platform string,
) (Parser, error) {

	name := strings.ToLower(
		strings.TrimSpace(platform),
	)

	parser, exists := r.parsers[name]

	if !exists {
		return nil, fmt.Errorf(
			"no parser registered for platform %q",
			platform,
		)
	}

	return parser, nil
}

func (r *ParserRegistry) RegisterDefaults() {

	r.Register(
		parsers.NewGitHubParser(),
	)

	r.Register(
		parsers.NewRedditParser(),
	)

	r.Register(
		parsers.NewXParser(),
	)
}
