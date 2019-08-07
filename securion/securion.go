package securion

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/blixenkrone/gopro/utils/logger"

	"github.com/blixenkrone/gopro/utils"
)

var log = logger.NewLogger()

// Client -
type Client struct {
	apiKey string
}

// Service -
type Service interface {
	GetPlans(string) ([]*Plan, error)
	GetPlansJSON(limit, period string) ([]*Plan, error)
	GetAPIKey() string
}

// NewClient returns new securion client with inherited interface
func NewClient() Service {
	apiKey := utils.LookupEnv("SECURION_KEY", "")

	return &Client{
		apiKey: apiKey,
	}
}

// Plan -
type Plan struct {
	ID       string            `json:"id"`
	Currency string            `json:"currency"`
	Interval string            `json:"interval"`
	Name     string            `json:"name"`
	Amount   int               `json:"amount"`
	Metadata map[string]string `json:"metadata"`
}

// Plans -
type Plans struct {
	List []*Plan `json:"list"`
}

// GetPlans -
func (c *Client) GetPlans(limit string) ([]*Plan, error) {
	var plans []*Plan
	url := fmt.Sprintf("https://api.securionpay.com/plans?limit=%s", limit)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&plans)
	return plans, nil
}

var StdPlans = []string{
	"plan_FX5jZjA8Orp2tVtqpl9YZAmk",
	"plan_AKtiNI1BNweN1PB1XefWdBd0",
	"plan_3dMlYjLKsLEArHP0aFCuXfmz",
}

// GetPlansJSON -
func (c *Client) GetPlansJSON(limit, period string) ([]*Plan, error) {
	var plans *Plans
	url := fmt.Sprintf("https://api.securionpay.com/plans?limit=%s", limit)
	res, err := c.doRequest(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	err = json.NewDecoder(res.Body).Decode(&plans)
	if err != nil {
		return nil, err
	}
	return plans.List, nil
}

func (c *Client) doRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.apiKey, "")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	log.Info("Did request to securion")
	return resp, nil
}

// GetAPIKey Returns APIKEY
func (c *Client) GetAPIKey() string {
	return c.apiKey
}
