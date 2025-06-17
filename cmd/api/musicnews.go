package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// NewsAPIResponse represents the response from News API
type NewsAPIResponse struct {
	Status       string    `json:"status"`
	TotalResults int       `json:"totalResults"`
	Articles     []Article `json:"articles"`
}

// Article represents a single news article
type Article struct {
	Source      Source `json:"source"`
	Author      string `json:"author"`
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	URLToImage  string `json:"urlToImage"`
	PublishedAt string `json:"publishedAt"`
	Content     string `json:"content"`
}

// Source represents the news source
type Source struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// NewsService handles all news-related operations
type NewsService struct {
	client *Optivet_Client
	config config
}

// NewNewsService creates a new news service instance
func NewNewsService(config config) *NewsService {
	client := NewClient(8*time.Second, 1) // 8s timeout, 1 retry - total max time ~16s for better UX
	return &NewsService{
		client: client,
		config: config,
	}
}

// FetchMusicNews fetches music news from News API
func (ns *NewsService) FetchMusicNews(newsType, country, genre string, limit int) (*NewsAPIResponse, error) {
	var endpoint string
	params := make(map[string]string)

	// Build music-specific query
	var musicQuery string
	baseQuery := "(music OR musician OR singer OR band OR album OR concert OR festival OR artist OR song OR Grammy OR Billboard)"

	if genre != "" {
		musicQuery = fmt.Sprintf("%s AND %s", baseQuery, genre)
	} else {
		musicQuery = baseQuery
	}

	// Exclude non-music content
	musicQuery += " AND NOT (politics OR sports OR business OR technology OR health OR science)"

	switch newsType {
	case "headlines":
		endpoint = "top-headlines"
		params["country"] = country
		params["q"] = musicQuery
		params["pageSize"] = strconv.Itoa(limit)
	default: // "everything"
		endpoint = "everything"
		params["q"] = musicQuery
		params["pageSize"] = strconv.Itoa(limit)
		params["sortBy"] = "publishedAt"
		params["language"] = "en"
	}

	// Add API key
	params["apiKey"] = ns.config.api.newsapi

	// Build the URL
	apiURL, err := buildAPIURL(ns.config.baseURLs.newsapi, endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("failed to build API URL: %w", err)
	}

	// Make the request using your HTTP client
	response, err := GETRequest[NewsAPIResponse](ns.client, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch news: %w", err)
	}

	// Check if News API returned an error
	if response.Status != "ok" {
		return nil, fmt.Errorf("news API error: %s", response.Status)
	}

	// Filter articles to ensure they are music-related
	response.Articles = ns.filterMusicArticles(response.Articles)
	response.TotalResults = len(response.Articles)

	return &response, nil
}

// filterMusicArticles filters articles to ensure they are music-related
func (ns *NewsService) filterMusicArticles(articles []Article) []Article {
	musicKeywords := []string{
		"music", "musician", "singer", "band", "album", "song", "artist", "concert",
		"festival", "grammy", "billboard", "spotify", "apple music", "streaming",
		"tour", "recording", "label", "producer", "rapper", "hip hop", "rock", "pop",
		"jazz", "classical", "country", "r&b", "electronic", "indie", "metal",
	}

	excludeKeywords := []string{
		"politics", "election", "government", "sports", "football", "basketball",
		"baseball", "soccer", "business merger", "stock market", "economy",
	}

	var filteredArticles []Article

	for _, article := range articles {
		// Combine title and description for keyword matching
		content := strings.ToLower(article.Title + " " + article.Description)

		// Check if article contains music-related keywords
		containsMusic := false
		for _, keyword := range musicKeywords {
			if strings.Contains(content, keyword) {
				containsMusic = true
				break
			}
		}

		// Check if article contains excluded keywords
		containsExcluded := false
		for _, keyword := range excludeKeywords {
			if strings.Contains(content, keyword) {
				containsExcluded = true
				break
			}
		}

		// Include article if it's music-related and doesn't contain excluded content
		if containsMusic && !containsExcluded {
			filteredArticles = append(filteredArticles, article)
		}
	}

	return filteredArticles
}

// getAllMusicalNews handles the request to fetch all musical news
func (app *application) getAllMusicalNews(w http.ResponseWriter, r *http.Request) {
	// Read query parameters using your existing method
	var input struct {
		limit    string
		country  string
		newsType string
		genre    string
	}

	qs := r.URL.Query()
	input.limit = app.readString(qs, "limit", "")
	input.country = app.readString(qs, "country", "us")
	input.newsType = app.readString(qs, "type", "everything")
	input.genre = app.readString(qs, "genre", "")

	// Set defaults and validate
	limit := 20 // Default number of articles

	// Parse limit parameter
	if input.limit != "" {
		if parsedLimit, err := strconv.Atoi(input.limit); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	// Create news service
	newsService := NewNewsService(app.config)

	// Fetch news
	response, err := newsService.FetchMusicNews(input.newsType, input.country, input.genre, limit)
	if err != nil {
		// Check for timeout or network errors
		if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "context deadline exceeded") {
			app.badRequestResponse(w, r, fmt.Errorf("news service is taking too long to respond. Please try again later"))
			return
		}

		// Check for API key errors or authentication issues
		if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "unauthorized") {
			app.serverErrorResponse(w, r, fmt.Errorf("news service authentication error"))
			return
		}

		// Check for rate limiting
		if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "rate limit") {
			app.badRequestResponse(w, r, fmt.Errorf("news service rate limit exceeded. Please try again later"))
			return
		}

		// For other errors, treat as server error
		app.serverErrorResponse(w, r, err)
		return
	}

	// Filter articles to ensure they are music-related
	filteredArticles := newsService.filterMusicArticles(response.Articles)

	// Update response with filtered articles
	response.Articles = filteredArticles

	// Return the response
	err = app.writeJSON(w, http.StatusOK, envelope{"news": response}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
