package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type CloudFlareError struct {
	Code    int
	Message string
}

type CloudFlareResponse struct {
	Success  bool
	Errors   []CloudFlareError
	Messages []string
	Result   json.RawMessage
}

type CloudFlareZoneItem struct {
	ID   string
	Name string
}

type CloudFlareSetting struct {
	ID         string
	Value      interface{}
	ModifiedOn string `json:"modified_on"`
	Editable   bool
}

type CloudFlareSettings []CloudFlareSetting

func (c CloudFlareSettings) ConfigItems() ConfigItems {
	config := ConfigItems{}
	for _, setting := range c {
		config[setting.ID] = setting.Value
	}

	return config
}

type CloudFlareRequestItem struct {
	Value interface{} `json:"value"`
}

type CloudFlare struct {
	Client *http.Client
	Query  *CloudFlareQuery
	log    *log.Logger
}

func (c *CloudFlare) Set(zone, id string, val interface{}) error {
	body, err := json.Marshal(&CloudFlareRequestItem{Value: val})
	if err != nil {
		return err
	}

	req, err := c.Query.NewRequestBody("PATCH",
		fmt.Sprintf("/zones/%s/settings/%s", zone, id),
		bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	_, err = c.MakeRequest(req)

	return err
}

func (c *CloudFlare) Settings(zoneID string) (CloudFlareSettings, error) {
	var settings CloudFlareSettings

	req, err := c.Query.NewRequest("GET", fmt.Sprintf("/zones/%s/settings", zoneID))
	if err != nil {
		return settings, err
	}

	response, err := c.MakeRequest(req)
	if err != nil {
		return settings, err
	}

	err = json.Unmarshal(response.Result, &settings)

	return settings, err
}

func (c *CloudFlare) Update(zone string, config ConfigItemsForUpdate) error {
	for key, vals := range config {
		c.log.Printf("Changing %q from %#v to %#v", key, vals.Current, vals.Expected)

		if err := c.Set(zone, key, vals.Expected); err != nil {
			return err
		}
	}

	return nil
}

func (c *CloudFlare) Zones() ([]CloudFlareZoneItem, error) {
	req, err := c.Query.NewRequest("GET", "/zones")
	if err != nil {
		return nil, err
	}

	response, err := c.MakeRequest(req)
	if err != nil {
		return nil, err
	}

	var zones []CloudFlareZoneItem
	err = json.Unmarshal(response.Result, &zones)

	return zones, err
}

func (c *CloudFlare) MakeRequest(request *http.Request) (*CloudFlareResponse, error) {
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

func NewCloudFlare(query *CloudFlareQuery, logger *log.Logger) *CloudFlare {
	return &CloudFlare{
		Client: &http.Client{},
		Query:  query,
		log:    logger,
	}
}
