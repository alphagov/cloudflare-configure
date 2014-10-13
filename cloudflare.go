package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type CloudFlareQuery struct {
	RootURL string
}

func (q *CloudFlareQuery) NewRequest(method, path string) (*http.Request, error) {
	url := fmt.Sprintf("%s%s", q.RootURL, path)
	return http.NewRequest(method, url, nil)
}

type CloudFlare struct {
	Client *http.Client
	Query  *CloudFlareQuery
}

func (c *CloudFlare) Settings(zoneID string) ([]CloudFlareConfigItem, error) {
	req, err := c.Query.NewRequest("GET", fmt.Sprintf("/zones/%s/settings", zoneID))
	if err != nil {
		return nil, err
	}

	response, err := c.makeRequest(req)
	if err != nil {
		return nil, err
	}

	var settings []CloudFlareConfigItem
	err = json.Unmarshal(response.Result, &settings)

	return settings, err
}

func (c *CloudFlare) Zones() ([]CloudFlareZoneItem, error) {
	req, err := c.Query.NewRequest("GET", "/zones")
	if err != nil {
		return nil, err
	}

	response, err := c.makeRequest(req)
	if err != nil {
		return nil, err
	}

	var zones []CloudFlareZoneItem
	err = json.Unmarshal(response.Result, &zones)

	return zones, err
}

func (c *CloudFlare) makeRequest(request *http.Request) (*CloudFlareResponse, error) {
	resp, err := c.Client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Didn't get 200 response, body: %s", body)
	}

	var response CloudFlareResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	if !response.Success || len(response.Errors) > 0 {
		return nil, fmt.Errorf("Response body indicated failure, response: %#v", response)
	}

	return &response, err
}

func NewCloudFlare(query *CloudFlareQuery) *CloudFlare {
	return &CloudFlare{
		Client: &http.Client{},
		Query:  query,
	}
}
