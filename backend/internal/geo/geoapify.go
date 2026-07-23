// Package geo proxeia a geocodificação de endereço via Geoapify — a chave
// nunca sai do backend (docs/migracao/01 §3.4).
package geo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const geoapifyBaseURL = "https://api.geoapify.com/v1/geocode/search"

type Resultado struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lng"`
	Preciso   bool    `json:"preciso"` // casou o número da casa
}

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{apiKey: apiKey, httpClient: &http.Client{Timeout: 10 * time.Second}}
}

type respostaGeoapify struct {
	Features []struct {
		Properties struct {
			Lat         float64 `json:"lat"`
			Lon         float64 `json:"lon"`
			HouseNumber string  `json:"housenumber"`
			ResultType  string  `json:"result_type"`
		} `json:"properties"`
	} `json:"features"`
}

// Geocode geocodifica um endereço no Brasil. Devolve nil (sem erro) se o
// Geoapify não achar nada — ausência de resultado não é uma falha de sistema.
func (c *Client) Geocode(ctx context.Context, endereco string) (*Resultado, error) {
	if endereco == "" {
		return nil, nil
	}
	q := url.Values{}
	q.Set("text", endereco)
	q.Set("filter", "countrycode:br")
	q.Set("limit", "1")
	q.Set("lang", "pt")
	q.Set("apiKey", c.apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, geoapifyBaseURL+"?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("geocodificar endereço: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("geocodificar endereço: HTTP %d", resp.StatusCode)
	}

	var out respostaGeoapify
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decodificar resposta do Geoapify: %w", err)
	}
	if len(out.Features) == 0 {
		return nil, nil
	}
	p := out.Features[0].Properties
	preciso := p.HouseNumber != "" || p.ResultType == "building"
	return &Resultado{Latitude: p.Lat, Longitude: p.Lon, Preciso: preciso}, nil
}
