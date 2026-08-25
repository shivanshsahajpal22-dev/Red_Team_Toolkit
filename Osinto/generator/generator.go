package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
)

// Target represents one item placed into the Redis queue.
//
// The worker can later use PlatformURL and Identifier to construct
// the actual request it needs to perform.
type Target struct {
	Platform    string `json:"platform"`
	PlatformURL string `json:"platform_url"`
	Identifier  string `json:"identifier"`
	Type        string `json:"type"` // username or email
}

// Seed contains the initial information supplied by the user.
type Seed struct {
	FirstName string
	LastName  string
	Alias     string
	Email     string
}

// Config contains generator configuration.
type Config struct {
	RedisAddr  string
	RedisDB    int
	QueueName  string
	Domains    []string
	Platforms  []Platform
}

// Platform describes the URL structure used by a worker.
type Platform struct {
	Name string
	URL  string
}

// ------------------------------------------------------------
// Main
// ------------------------------------------------------------

func main() {
	ctx := context.Background()

	config := loadConfig()

	log.Printf("Starting OSINTO generator")
	log.Printf("Redis: %s", config.RedisAddr)
	log.Printf("Queue: %s", config.QueueName)

	redisClient := redis.NewClient(&redis.Options{
		Addr: config.RedisAddr,
		DB:   config.RedisDB,
	})

	defer redisClient.Close()

	// Verify Redis connectivity.
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Redis connection failed: %v", err)
	}

	seed := loadSeed()

	// Generate username permutations.
	usernames := generateUsernames(seed)

	log.Printf("Generated %d unique usernames", len(usernames))

	// Generate email permutations.
	emails := generateEmails(usernames, seed, config.Domains)

	log.Printf("Generated %d unique emails", len(emails))

	// Push username targets.
	usernameCount := 0

	for _, username := range usernames {
		for _, platform := range config.Platforms {

			target := Target{
				Platform:    platform.Name,
				PlatformURL: platform.URL,
				Identifier:  username,
				Type:        "username",
			}

			if err := pushTarget(ctx, redisClient, config.QueueName, target); err != nil {
				log.Fatalf("Failed to queue username target: %v", err)
			}

			usernameCount++
		}
	}

	// Push email targets.
	emailCount := 0

	for _, email := range emails {

		target := Target{
			Platform:    "email",
			PlatformURL: "",
			Identifier:  email,
			Type:        "email",
		}

		if err := pushTarget(ctx, redisClient, config.QueueName, target); err != nil {
			log.Fatalf("Failed to queue email target: %v", err)
		}

		emailCount++
	}

	log.Printf(
		"Generator finished: %d username targets + %d email targets",
		usernameCount,
		emailCount,
	)
}

// ------------------------------------------------------------
// Username generation
// ------------------------------------------------------------

func generateUsernames(seed Seed) []string {
	seen := make(map[string]struct{})

	add := func(value string) {
		value = normalizeUsername(value)

		if value == "" {
			return
		}

		if len(value) > 64 {
			return
		}

		seen[value] = struct{}{}
	}

	first := normalizeUsername(seed.FirstName)
	last := normalizeUsername(seed.LastName)
	alias := normalizeUsername(seed.Alias)

	// Original alias.
	add(alias)

	// First-name / last-name combinations.
	if first != "" && last != "" {

		add(first + last)
		add(last + first)

		add(first + "_" + last)
		add(last + "_" + first)

		add(first + "." + last)
		add(last + "." + first)

		add(first + "-" + last)
		add(last + "-" + first)

		add(first + " " + last)
		add(last + " " + first)

		// Initial combinations.
		add(string(first[0]) + last)
		add(first + string(last[0]))

		add(string(first[0]) + "_" + last)
		add(first + "_" + string(last[0]))

		add(string(first[0]) + "." + last)
		add(first + "." + string(last[0]))
	}

	// Useful numeric suffixes.
	baseValues := []string{
		first,
		last,
		alias,
		first + last,
		last + first,
	}

	suffixes := []string{
		"1",
		"01",
		"99",
		"123",
		"007",
		"1234",
	}

	for _, base := range baseValues {

		if base == "" {
			continue
		}

		for _, suffix := range suffixes {
			add(base + suffix)
			add(base + "_" + suffix)
			add(base + "." + suffix)
		}
	}

	// Prefix patterns.
	prefixes := []string{
		"the",
		"real",
		"official",
		"its",
		"iam",
	}

	for _, prefix := range prefixes {

		for _, base := range baseValues {

			if base == "" {
				continue
			}

			add(prefix + base)
			add(prefix + "_" + base)
			add(prefix + "." + base)
		}
	}

	// Convert map into slice.
	result := make([]string, 0, len(seen))

	for username := range seen {
		result = append(result, username)
	}

	return result
}

// ------------------------------------------------------------
// Email generation
// ------------------------------------------------------------

func generateEmails(
	usernames []string,
	seed Seed,
	domains []string,
) []string {

	seen := make(map[string]struct{})

	add := func(email string) {
		email = strings.ToLower(strings.TrimSpace(email))

		if email == "" {
			return
		}

		seen[email] = struct{}{}
	}

	// Preserve an explicitly supplied email.
	if seed.Email != "" {
		add(seed.Email)
	}

	for _, username := range usernames {

		for _, domain := range domains {

			domain = strings.TrimSpace(
				strings.ToLower(domain),
			)

			if domain == "" {
				continue
			}

			add(username + "@" + domain)
		}
	}

	result := make([]string, 0, len(seen))

	for email := range seen {
		result = append(result, email)
	}

	return result
}

// ------------------------------------------------------------
// Redis
// ------------------------------------------------------------

func pushTarget(
	ctx context.Context,
	client *redis.Client,
	queue string,
	target Target,
) error {

	payload, err := json.Marshal(target)

	if err != nil {
		return fmt.Errorf(
			"serialize target: %w",
			err,
		)
	}

	// Architecture specifies LPUSH for queue population.
	if err := client.LPush(
		ctx,
		queue,
		string(payload),
	).Err(); err != nil {

		return fmt.Errorf(
			"LPUSH target: %w",
			err,
		)
	}

	return nil
}

// ------------------------------------------------------------
// Input
// ------------------------------------------------------------

func loadSeed() Seed {

	seed := Seed{
		FirstName: os.Getenv("FIRST_NAME"),
		LastName:  os.Getenv("LAST_NAME"),
		Alias:     os.Getenv("ALIAS"),
		Email:     os.Getenv("EMAIL"),
	}

	if seed.FirstName == "" &&
		seed.LastName == "" &&
		seed.Alias == "" &&
		seed.Email == "" {

		log.Println(
			"Warning: no seed values supplied.",
		)
	}

	return seed
}

// ------------------------------------------------------------
// Configuration
// ------------------------------------------------------------

func loadConfig() Config {

	redisAddr := os.Getenv("REDIS_ADDR")

	if redisAddr == "" {
		redisAddr = "redis:6379"
	}

	queueName := os.Getenv("REDIS_QUEUE")

	if queueName == "" {
		queueName = "osinto:targets"
	}

	redisDB := 0

	if value := os.Getenv("REDIS_DB"); value != "" {

		if parsed, err := strconv.Atoi(value); err == nil {
			redisDB = parsed
		}
	}

	domains := parseCommaSeparated(
		os.Getenv("EMAIL_DOMAINS"),
	)

	// Sensible defaults if no custom domains were supplied.
	if len(domains) == 0 {
		domains = []string{
			"gmail.com",
			"yahoo.com",
			"outlook.com",
			"hotmail.com",
		}
	}

	platforms := []Platform{
		{
			Name: "github",
			URL:  "https://github.com/{username}",
		},
		{
			Name: "reddit",
			URL:  "https://www.reddit.com/user/{username}",
		},
		{
			Name: "x",
			URL:  "https://x.com/{username}",
		},
	}

	return Config{
		RedisAddr: queueNameRedisAddr(redisAddr),
		RedisDB:   redisDB,
		QueueName: queueName,
		Domains:   domains,
		Platforms: platforms,
	}
}

func queueNameRedisAddr(addr string) string {
	return addr
}

// ------------------------------------------------------------
// Helpers
// ------------------------------------------------------------

func normalizeUsername(value string) string {

	value = strings.ToLower(
		strings.TrimSpace(value),
	)

	// Replace whitespace with separators.
	value = strings.ReplaceAll(
		value,
		" ",
		"_",
	)

	// Remove characters that don't belong in a
	// conventional username candidate.
	var builder strings.Builder

	for _, r := range value {

		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)

		case r >= '0' && r <= '9':
			builder.WriteRune(r)

		case r == '_':
			builder.WriteRune(r)

		case r == '.':
			builder.WriteRune(r)

		case r == '-':
			builder.WriteRune(r)
		}
	}

	return strings.Trim(
		builder.String(),
		"_.-",
	)
}

func parseCommaSeparated(value string) []string {

	if strings.TrimSpace(value) == "" {
		return nil
	}

	parts := strings.Split(value, ",")

	result := make([]string, 0, len(parts))

	for _, part := range parts {

		part = strings.TrimSpace(part)

		if part != "" {
			result = append(result, part)
		}
	}

	return result
}
