package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/net/html"
	"osinto/worker/model"
)

type Worker struct {
	Redis      *redis.Client
	Queue      string
	HTTPClient *http.Client
	UserAgent  string
}

func main() {
	ctx := context.Background()

	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	queue := getEnv("REDIS_QUEUE", "osinto:targets")

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	defer redisClient.Close()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Redis connection failed: %v", err)
	}

	worker := &Worker{
		Redis: redisClient,
		Queue: queue,

		HTTPClient: &http.Client{
			Timeout: 15 * time.Second,
		},

		UserAgent: getEnv(
			"USER_AGENT",
			"OSINTO/1.0",
		),
	}

	log.Printf("OSINTO worker started")
	log.Printf("Redis: %s", redisAddr)
	log.Printf("Queue: %s", queue)

	worker.Run(ctx)
}

func (w *Worker) Run(ctx context.Context) {
	for {
		// BRPOP blocks until an item becomes available.
		result, err := w.Redis.BRPop(
			ctx,
			0,
			w.Queue,
		).Result()

		if err != nil {
			if ctx.Err() != nil {
				return
			}

			log.Printf("Redis BRPOP error: %v", err)
			time.Sleep(time.Second)
			continue
		}

		if len(result) != 2 {
			continue
		}

		payload := result[1]

		var target Target

		if err := json.Unmarshal(
			[]byte(payload),
			&target,
		); err != nil {
			log.Printf(
				"Invalid target payload: %v",
				err,
			)
			continue
		}

		log.Printf(
			"Processing %s:%s",
			target.Platform,
			target.Identifier,
		)

		if err := w.ProcessTarget(ctx, target); err != nil {
			log.Printf(
				"Target processing failed: %v",
				err,
			)
		}
	}
}

func (w *Worker) ProcessTarget(
	ctx context.Context,
	target Target,
) error {

	if target.Type != "username" {
		log.Printf(
			"Skipping unsupported target type: %s",
			target.Type,
		)
		return nil
	}

	url := strings.ReplaceAll(
		target.PlatformURL,
		"{username}",
		target.Identifier,
	)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		url,
		nil,
	)

	if err != nil {
		return fmt.Errorf(
			"create request: %w",
			err,
		)
	}

	req.Header.Set(
		"User-Agent",
		w.UserAgent,
	)

	req.Header.Set(
		"Accept",
		"text/html,application/xhtml+xml",
	)

	resp, err := w.HTTPClient.Do(req)

	if err != nil {
		return fmt.Errorf(
			"HTTP request: %w",
			err,
		)
	}

	defer resp.Body.Close()

	switch resp.StatusCode {

	case http.StatusNotFound:
		log.Printf(
			"[%s] %s -> account does not exist",
			target.Platform,
			target.Identifier,
		)
		return nil

	case http.StatusGone:
		log.Printf(
			"[%s] %s -> account is gone",
			target.Platform,
			target.Identifier,
		)
		return nil

	case http.StatusOK:
		// Continue below.

	default:
		log.Printf(
			"[%s] %s -> HTTP %d",
			target.Platform,
			target.Identifier,
			resp.StatusCode,
		)

		return nil
	}

	body, err := io.ReadAll(
		io.LimitReader(resp.Body, 5*1024*1024),
	)

	if err != nil {
		return fmt.Errorf(
			"read response: %w",
			err,
		)
	}

	profile, err := parseProfile(
		target,
		string(body),
	)

	if err != nil {
		return fmt.Errorf(
			"parse profile: %w",
			err,
		)
	}

	log.Printf(
		"FOUND [%s] %s",
		profile.Platform,
		profile.Identifier,
	)

	log.Printf(
		"  Display Name: %s",
		profile.DisplayName,
	)

	log.Printf(
		"  Bio: %s",
		profile.Bio,
	)

	log.Printf(
		"  Location: %s",
		profile.Location,
	)

	log.Printf(
		"  Links: %d",
		len(profile.Links),
	)

	// PostgreSQL storage will be added here.
	//
	// saveProfile(profile)

	return nil
}

func parseProfile(
	target Target,
	body string,
) (Profile, error) {

	doc, err := html.Parse(
		strings.NewReader(body),
	)

	if err != nil {
		return Profile{}, err
	}

	profile := Profile{
		Platform:    target.Platform,
		Identifier:  target.Identifier,
		URL: strings.ReplaceAll(
			target.PlatformURL,
			"{username}",
			target.Identifier,
		),
		Links: []string{},
	}

	extractDocument(doc, &profile)

	return profile, nil
}

func extractDocument(
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

		extractDocument(
			child,
			profile,
		)
	}
}

func nodeText(node *html.Node) string {

	var builder strings.Builder

	var walk func(*html.Node)

	walk = func(n *html.Node) {

		if n.Type == html.TextNode {
			builder.WriteString(n.Data)
			builder.WriteString(" ")
		}

		for child := n.FirstChild;
			child != nil;
			child = child.NextSibling {

			walk(child)
		}
	}

	walk(node)

	return strings.Join(
		strings.Fields(
			builder.String(),
		),
		" ",
	)
}

func getEnv(
	key string,
	fallback string,
) string {

	value := strings.TrimSpace(
		os.Getenv(key),
	)

	if value == "" {
		return fallback
	}

	return value
}
