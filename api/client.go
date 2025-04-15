package api

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type SuggestClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewSuggestClient(apiKey string) *SuggestClient {
	return &SuggestClient{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *SuggestClient) GetSuggestions(query string, lat, lon, radius float64) ([]byte, error) {
	urlStr := fmt.Sprintf(
		"https://suggest-maps.yandex.ru/v1/suggest?apikey=%s&text=%s&ll=%.6f,%.6f&spn=%.6f,%.6f&strict_bounds=1&print_address=1",
		c.apiKey,
		url.QueryEscape(query),
		lon,
		lat,
		radius,
		radius,
	)

	resp, err := c.httpClient.Get(urlStr)
	if err != nil {
		return nil, fmt.Errorf("suggest request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

type GeocoderClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewGeocoderClient(apiKey string) *GeocoderClient {
	return &GeocoderClient{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *GeocoderClient) Geocode(lat, lon float64) ([]byte, error) {
	urlStr := fmt.Sprintf(
		"https://geocode-maps.yandex.ru/1.x/?apikey=%s&format=json&geocode=%.6f,%.6f",
		c.apiKey,
		lon,
		lat,
	)

	resp, err := c.httpClient.Get(urlStr)
	if err != nil {
		return nil, fmt.Errorf("geocode request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (c *GeocoderClient) ReverseGeocode(address string) ([]byte, error) {
	urlStr := fmt.Sprintf(
		"https://geocode-maps.yandex.ru/1.x/?apikey=%s&format=json&geocode=%s",
		c.apiKey,
		url.QueryEscape(address),
	)

	resp, err := c.httpClient.Get(urlStr)
	if err != nil {
		return nil, fmt.Errorf("reverse geocode request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}