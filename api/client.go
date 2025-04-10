package api

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

type MapsClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewMapsClient(apiKey string) *MapsClient {
	return &MapsClient{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *MapsClient) GetLocation(lat, lon float64) ([]byte, error) {
	latStr := fmt.Sprintf("%.6f", lat)
	lngStr := fmt.Sprintf("%.6f", lon)

	url := fmt.Sprintf("https://geocode-maps.yandex.ru/1.x/?apikey=%s&format=json&geocode=%s,%s",
		c.apiKey,
		lngStr,
		latStr,
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	return body, nil
}