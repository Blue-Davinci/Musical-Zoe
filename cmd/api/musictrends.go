package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// LastFMResponse represents the response from Last.fm API
type LastFMResponse struct {
	Tracks TracksContainer `json:"tracks"`
}

// TracksContainer wraps the tracks array
type TracksContainer struct {
	Track []Track `json:"track"`
}

// Track represents a single track from Last.fm
type Track struct {
	Name       string     `json:"name"`
	Duration   string     `json:"duration"`
	PlayCount  string     `json:"playcount"`
	Listeners  string     `json:"listeners"`
	MBID       string     `json:"mbid"`
	URL        string     `json:"url"`
	Streamable Streamable `json:"streamable"`
	Artist     Artist     `json:"artist"`
	Image      []Image    `json:"image"`
}

// Streamable represents streaming information
type Streamable struct {
	Text      string `json:"#text"`
	FullTrack string `json:"fulltrack"`
}

// Artist represents artist information
type Artist struct {
	Name string `json:"name"`
	MBID string `json:"mbid"`
	URL  string `json:"url"`
}

// Image represents track artwork in different sizes
type Image struct {
	Text string `json:"#text"`
	Size string `json:"size"`
}

// TrendsService handles all Last.fm trends-related operations
type TrendsService struct {
	client *Optivet_Client
	config config
}

// NewTrendsService creates a new trends service instance
func NewTrendsService(config config) *TrendsService {
	client := NewClient(8*time.Second, 1) // 8s timeout, 1 retry - total max time ~16s for better UX
	return &TrendsService{
		client: client,
		config: config,
	}
}

// FetchTopTracks fetches trending tracks from Last.fm
func (ts *TrendsService) FetchTopTracks(limit int, period string) (*LastFMResponse, error) {
	params := make(map[string]string)

	// Set method and API key
	params["method"] = "chart.gettoptracks"
	params["api_key"] = ts.config.api.lastfm
	params["format"] = "json"
	params["limit"] = strconv.Itoa(limit)

	// Add period if specified (7day, 1month, 3month, 6month, 12month, overall)
	if period != "" {
		params["period"] = period
	}

	// Build the URL using the helper function
	apiURL, err := buildAPIURL(ts.config.baseURLs.lastfm, "", params)
	if err != nil {
		return nil, fmt.Errorf("failed to build API URL: %w", err)
	}

	// Make the request using your HTTP client
	response, err := GETRequest[LastFMResponse](ts.client, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch trends: %w", err)
	}

	return &response, nil
}

// FetchTopArtists fetches trending artists from Last.fm
func (ts *TrendsService) FetchTopArtists(limit int, period string) (*LastFMArtistsResponse, error) {
	params := make(map[string]string)

	params["method"] = "chart.gettopartists"
	params["api_key"] = ts.config.api.lastfm
	params["format"] = "json"
	params["limit"] = strconv.Itoa(limit)

	if period != "" {
		params["period"] = period
	}

	// Build the URL
	apiURL, err := buildAPIURL(ts.config.baseURLs.lastfm, "", params)
	if err != nil {
		return nil, fmt.Errorf("failed to build API URL: %w", err)
	}

	// Make the request
	response, err := GETRequest[LastFMArtistsResponse](ts.client, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch top artists: %w", err)
	}

	return &response, nil
}

// LastFMArtistsResponse represents the response for top artists
type LastFMArtistsResponse struct {
	Artists ArtistsContainer `json:"artists"`
}

// ArtistsContainer wraps the artists array
type ArtistsContainer struct {
	Artist []ArtistInfo `json:"artist"`
}

// ArtistInfo represents detailed artist information
type ArtistInfo struct {
	Name      string  `json:"name"`
	PlayCount string  `json:"playcount"`
	Listeners string  `json:"listeners"`
	MBID      string  `json:"mbid"`
	URL       string  `json:"url"`
	Image     []Image `json:"image"`
}

// getAllMusicTrends handles the request to fetch music trends
func (app *application) getAllMusicTrends(w http.ResponseWriter, r *http.Request) {
	// Read query parameters
	var input struct {
		limit  string
		period string
		tType  string // tracks or artists
	}

	qs := r.URL.Query()
	input.limit = app.readString(qs, "limit", "")
	input.period = app.readString(qs, "period", "")
	input.tType = app.readString(qs, "type", "tracks")

	// Set defaults and validate
	limit := 50 // Default number of items
	if input.limit != "" {
		if parsedLimit, err := strconv.Atoi(input.limit); err == nil && parsedLimit > 0 && parsedLimit <= 200 {
			limit = parsedLimit
		}
	}

	// Validate period
	validPeriods := map[string]bool{
		"7day": true, "1month": true, "3month": true,
		"6month": true, "12month": true, "overall": true,
	}
	if input.period != "" && !validPeriods[input.period] {
		app.badRequestResponse(w, r, fmt.Errorf("invalid period. Valid periods: 7day, 1month, 3month, 6month, 12month, overall"))
		return
	}

	// Create trends service
	trendsService := NewTrendsService(app.config)

	// Fetch data based on type
	switch input.tType {
	case "artists":
		response, err := trendsService.FetchTopArtists(limit, input.period)
		if err != nil {
			// Check for timeout or network errors
			if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "context deadline exceeded") {
				app.badRequestResponse(w, r, fmt.Errorf("trends service is taking too long to respond. Please try again later"))
				return
			}

			// Check for API key errors
			if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "unauthorized") || strings.Contains(err.Error(), "invalid api key") {
				app.serverErrorResponse(w, r, fmt.Errorf("trends service authentication error"))
				return
			}

			// Check for rate limiting
			if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "rate limit") {
				app.badRequestResponse(w, r, fmt.Errorf("trends service rate limit exceeded. Please try again later"))
				return
			}

			app.serverErrorResponse(w, r, err)
			return
		}
		err = app.writeJSON(w, http.StatusOK, envelope{"trends": response}, nil)
		if err != nil {
			app.serverErrorResponse(w, r, err)
		}
	default: // "tracks"
		response, err := trendsService.FetchTopTracks(limit, input.period)
		if err != nil {
			// Check for timeout or network errors
			if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "context deadline exceeded") {
				app.badRequestResponse(w, r, fmt.Errorf("trends service is taking too long to respond. Please try again later"))
				return
			}

			// Check for API key errors
			if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "unauthorized") || strings.Contains(err.Error(), "invalid api key") {
				app.serverErrorResponse(w, r, fmt.Errorf("trends service authentication error"))
				return
			}

			// Check for rate limiting
			if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "rate limit") {
				app.badRequestResponse(w, r, fmt.Errorf("trends service rate limit exceeded. Please try again later"))
				return
			}

			app.serverErrorResponse(w, r, err)
			return
		}
		err = app.writeJSON(w, http.StatusOK, envelope{"trends": response}, nil)
		if err != nil {
			app.serverErrorResponse(w, r, err)
		}
	}
}
