package duffel

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.duffel.com"

type Client struct {
	HTTPClient *http.Client
	BaseURL    string
	APIToken   string
}

type FlightOfferResponse struct {
	Data DuffelOfferRequestData `json:"data"`
}

type DuffelOfferRequestData struct {
	ID     string        `json:"id"`
	Offers []DuffelOffer `json:"offers"`
}

type DuffelOffer struct {
	ID            string `json:"id"`
	TotalAmount   string `json:"total_amount"`
	TotalCurrency string `json:"total_currency"`
}

type DuffelErrorResponse struct {
	Errors []DuffelError `json:"errors"`
}

type DuffelError struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

func NewClient(apiToken string) (*Client, error) {

	if strings.TrimSpace(apiToken) == "" {
		return nil, errors.New("API token is required")
	}
	return &Client{
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		BaseURL:  defaultBaseURL,
		APIToken: apiToken,
	}, nil
}

func (c *Client) newRequest(
	method string,
	endpoint string,
	body io.Reader,
) (*http.Request, error) {
	req, err := http.NewRequest(method, c.BaseURL+endpoint, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Duffel-Version", "v2")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}
func (c *Client) do(req *http.Request) (*http.Response, error) {
	return c.HTTPClient.Do(req)
}
func (c *Client) handleResponse(resp *http.Response) (*FlightOfferResponse, error) {
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		body, err := io.ReadAll(resp.Body)
        if err != nil {
            return nil, err
        }

		var flightOfferResp FlightOfferResponse
		err = json.Unmarshal(body, &flightOfferResp)
		if err != nil {
			return nil, err
		}
		return &flightOfferResp, nil
	} else {
		var duffelErrResp DuffelErrorResponse
		err := json.NewDecoder(resp.Body).Decode(&duffelErrResp)
		if err != nil {
			return nil, fmt.Errorf(
				"Duffel API returned %s; failed to decode error response: %w",
				resp.Status,
				err,
			)
		}
		apiErr := duffelErrResp.Errors[0]
		return nil, fmt.Errorf(
			"Duffel API returned %s: %s - %s",
			resp.Status,
			apiErr.Title,
			apiErr.Message,
		)
	}

}

func (c *Client) SearchFlights(ctx context.Context, req DuffelSearchParams) (*FlightOfferResponse, error) {
	searchRequest := buildSearchRequest(req)

	reqBody, err := json.Marshal(searchRequest)
	if err != nil {
		return nil, err
	}
	httpReq, err := c.newRequest("POST", "/air/offer_requests", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	resp, err := c.do(httpReq)
	if err != nil {
		return nil, err
	}
	return c.handleResponse(resp)
}
