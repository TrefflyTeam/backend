package geoservice

import (
	db "treffly/db/sqlc"
)

type Service struct {
	store      db.Store
	geocodeClient *GeocoderClient
	suggestClient *SuggestClient
}

func New(store db.Store, geocodeClient *GeocoderClient, suggestClient *SuggestClient) *Service {
	return &Service{
		store:      store,
		geocodeClient: geocodeClient,
		suggestClient: suggestClient,
	}
}

func (s *Service) Geocode(lat, lon float64) ([]byte, error) {
	body, err := s.geocodeClient.Geocode(lat, lon)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (s *Service) ReverseGeocode(address string) ([]byte, error) {
	body, err := s.geocodeClient.ReverseGeocode(address)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (s *Service) GetSuggestions(query string, lat, lon, radius float64) ([]byte, error) {
	body, err := s.suggestClient.GetSuggestions(query, lat, lon, radius)
	if err != nil {
		return nil, err
	}

	return body, nil
}

