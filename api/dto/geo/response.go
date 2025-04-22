package geodto

type SuggestResponse struct {
	Results []struct {
		Title struct {
			Text string `json:"text"`
		} `json:"title"`
		Address struct {
			FormattedAddress string `json:"formatted_address"`
			Components []struct {
				Name string   `json:"name"`
				Kind []string `json:"kind"`
			} `json:"component"`
		} `json:"address"`
	} `json:"results"`
}

type SuggestItem struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Address string `json:"address"`
}

type GeoResponse struct {
	Response struct {
		GeoObjectCollection struct {
			FeatureMember []struct {
				GeoObject struct {
					MetaDataProperty struct {
						GeocoderMetaData struct {
							Text    string `json:"text"`
							Address struct {
								Formatted string `json:"formatted"`
							} `json:"Address"`
						} `json:"GeocoderMetaData"`
					} `json:"metaDataProperty"`
					Point struct {
						Pos string `json:"pos"`
					} `json:"Point"`
				} `json:"GeoObject"`
			} `json:"featureMember"`
		} `json:"GeoObjectCollection"`
	} `json:"response"`
}

type LocationResult struct {
	Address string
	Lat     float64
	Lon     float64
}
