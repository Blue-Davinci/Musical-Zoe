package main

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// LyricsResponse represents the response from Lyrics.ovh API
type LyricsResponse struct {
	Lyrics string `json:"lyrics,omitempty"`
	Error  string `json:"error,omitempty"`
}

// ProcessedLyricsResponse represents the cleaned and processed lyrics response
type ProcessedLyricsResponse struct {
	Artist        string   `json:"artist"`
	Title         string   `json:"title"`
	Lyrics        string   `json:"lyrics"`
	CleanedLyrics []string `json:"cleaned_lyrics"`
	LinesCount    int      `json:"lines_count"`
	VersesCount   int      `json:"verses_count"`
	HasChorus     bool     `json:"has_chorus"`
	WordCount     int      `json:"word_count"`
	Source        string   `json:"source"`
	Status        string   `json:"status"`
}

// LyricsService handles all lyrics-related operations
type LyricsService struct {
	client *Optivet_Client
	config config
}

// NewLyricsService creates a new lyrics service instance
func NewLyricsService(config config) *LyricsService {
	client := NewClient(5*time.Second, 1) // 5s timeout, 1 retry - total max time ~10s for better UX
	return &LyricsService{
		client: client,
		config: config,
	}
}

// FetchLyrics fetches lyrics for a given artist and song title
func (ls *LyricsService) FetchLyrics(artist, title string) (*ProcessedLyricsResponse, error) {
	// URL encode the artist and title
	encodedArtist := url.QueryEscape(strings.TrimSpace(artist))
	encodedTitle := url.QueryEscape(strings.TrimSpace(title))

	// Build the URL manually since lyrics.ovh uses path parameters
	apiURL := fmt.Sprintf("%s/%s/%s", ls.config.baseURLs.lyrics, encodedArtist, encodedTitle)

	// Make the request
	response, err := GETRequest[LyricsResponse](ls.client, apiURL, nil)
	if err != nil {
		// Check for timeout errors
		if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "context deadline exceeded") {
			return nil, fmt.Errorf("request timeout: the lyrics service is taking too long to respond. Please try again later")
		}

		// Check if it's a 404 error (lyrics not found) - updated to match actual error format
		if strings.Contains(err.Error(), "non-2xx response code: 404") {
			return &ProcessedLyricsResponse{
				Artist: artist,
				Title:  title,
				Status: "not_found",
				Source: "lyrics.ovh",
			}, nil
		}

		// Check for other network errors
		if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "no such host") {
			return nil, fmt.Errorf("network error: unable to connect to lyrics service. Please try again later")
		}

		return nil, fmt.Errorf("failed to fetch lyrics: %w", err)
	}

	// Check if API returned an error
	if response.Error != "" {
		return &ProcessedLyricsResponse{
			Artist: artist,
			Title:  title,
			Status: "not_found",
			Source: "lyrics.ovh",
		}, nil
	}

	// Process and clean the lyrics
	processedResponse := ls.processLyrics(artist, title, response.Lyrics)
	return processedResponse, nil
}

// processLyrics cleans and processes raw lyrics text
func (ls *LyricsService) processLyrics(artist, title, rawLyrics string) *ProcessedLyricsResponse {
	if rawLyrics == "" {
		return &ProcessedLyricsResponse{
			Artist: artist,
			Title:  title,
			Status: "empty",
			Source: "lyrics.ovh",
		}
	}

	// Clean the lyrics
	cleanedLyrics := ls.cleanLyricsText(rawLyrics)
	lyricsLines := ls.splitIntoLines(cleanedLyrics)

	// Analyze the lyrics
	wordCount := ls.countWords(cleanedLyrics)
	versesCount := ls.countVerses(lyricsLines)
	hasChorus := ls.detectChorus(cleanedLyrics)

	return &ProcessedLyricsResponse{
		Artist:        artist,
		Title:         title,
		Lyrics:        cleanedLyrics,
		CleanedLyrics: lyricsLines,
		LinesCount:    len(lyricsLines),
		VersesCount:   versesCount,
		HasChorus:     hasChorus,
		WordCount:     wordCount,
		Source:        "lyrics.ovh",
		Status:        "found",
	}
}

// cleanLyricsText cleans and formats the raw lyrics text
func (ls *LyricsService) cleanLyricsText(text string) string {
	// Replace \r\n and \n with consistent line breaks
	cleaned := strings.ReplaceAll(text, "\r\n", "\n")
	cleaned = strings.ReplaceAll(cleaned, "\r", "\n")

	// Remove excessive empty lines (more than 2 consecutive)
	re := regexp.MustCompile(`\n{3,}`)
	cleaned = re.ReplaceAllString(cleaned, "\n\n")

	// Trim whitespace from each line
	lines := strings.Split(cleaned, "\n")
	var trimmedLines []string
	for _, line := range lines {
		trimmedLines = append(trimmedLines, strings.TrimSpace(line))
	}

	// Join back and trim overall
	cleaned = strings.Join(trimmedLines, "\n")
	cleaned = strings.TrimSpace(cleaned)

	return cleaned
}

// splitIntoLines splits lyrics into non-empty lines
func (ls *LyricsService) splitIntoLines(lyrics string) []string {
	lines := strings.Split(lyrics, "\n")
	var nonEmptyLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			nonEmptyLines = append(nonEmptyLines, trimmed)
		}
	}

	return nonEmptyLines
}

// countWords counts the total number of words in lyrics
func (ls *LyricsService) countWords(lyrics string) int {
	// Remove punctuation and split by whitespace
	re := regexp.MustCompile(`[^\w\s]`)
	cleaned := re.ReplaceAllString(lyrics, " ")
	words := strings.Fields(cleaned)
	return len(words)
}

// countVerses estimates the number of verses (groups of lines separated by empty lines)
func (ls *LyricsService) countVerses(lines []string) int {
	if len(lines) == 0 {
		return 0
	}

	// For simplified counting, estimate based on line breaks in original text
	// This is a basic heuristic since we've already removed empty lines
	// We'll estimate 1 verse per 4-6 lines on average
	return (len(lines) / 4) + 1
}

// detectChorus detects if the lyrics likely contain a chorus (repeated sections)
func (ls *LyricsService) detectChorus(lyrics string) bool {
	// Simple heuristic: look for common chorus indicators
	lowerLyrics := strings.ToLower(lyrics)
	chorusIndicators := []string{"chorus", "hook", "refrain"}

	for _, indicator := range chorusIndicators {
		if strings.Contains(lowerLyrics, indicator) {
			return true
		}
	}

	// Also check for repeated phrases (very basic)
	lines := strings.Split(lyrics, "\n")
	lineMap := make(map[string]int)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) > 10 { // Only consider substantial lines
			lineMap[trimmed]++
			if lineMap[trimmed] > 1 {
				return true // Found repeated line, likely a chorus
			}
		}
	}

	return false
}

// getLyrics handles the request to fetch lyrics for a song
func (app *application) getLyrics(w http.ResponseWriter, r *http.Request) {
	// Read query parameters
	var input struct {
		artist string
		title  string
		format string
	}

	qs := r.URL.Query()
	input.artist = app.readString(qs, "artist", "")

	// Accept both "title" and "song" as aliases for the song title
	input.title = app.readString(qs, "title", "")
	if input.title == "" {
		input.title = app.readString(qs, "song", "")
	}

	input.format = app.readString(qs, "format", "processed")

	// Validate required parameters
	if input.artist == "" {
		app.badRequestResponse(w, r, fmt.Errorf("artist parameter is required"))
		return
	}

	if input.title == "" {
		app.badRequestResponse(w, r, fmt.Errorf("title or song parameter is required"))
		return
	}

	// Create lyrics service
	lyricsService := NewLyricsService(app.config)

	// Fetch lyrics
	response, err := lyricsService.FetchLyrics(input.artist, input.title)
	if err != nil {
		// Check for timeout errors
		if strings.Contains(err.Error(), "request timeout") {
			app.badRequestResponse(w, r, err)
			return
		}

		// Check for network errors
		if strings.Contains(err.Error(), "network error") {
			app.badRequestResponse(w, r, err)
			return
		}

		// For other errors, treat as server error
		app.serverErrorResponse(w, r, err)
		return
	}

	// Check if lyrics were not found
	if response.Status == "not_found" {
		app.notFoundResponse(w, r)
		return
	}

	// Return response based on format
	if input.format == "raw" && response.Status == "found" {
		// Return just the raw lyrics for simple use cases
		err = app.writeJSON(w, http.StatusOK, envelope{
			"artist": response.Artist,
			"title":  response.Title,
			"lyrics": response.Lyrics,
			"source": response.Source,
		}, nil)
	} else {
		// Return full processed response
		err = app.writeJSON(w, http.StatusOK, envelope{"lyrics": response}, nil)
	}

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
