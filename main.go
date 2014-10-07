package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type Config map[string]interface{}

type Response struct {
	Success  bool
	Errors   []string
	Messages []string
	Result   []ResponseSetting
}

type ResponseSetting struct {
	ID         string
	Value      interface{}
	ModifiedOn string `json:"modified_on"`
	Editable   bool
}

const rootURL = "https://api.cloudflare.com/v4"

var httpClient = &http.Client{}

func main() {
	var (
		authEmail  = flag.String("email", "", "Authentication email address")
		authKey    = flag.String("key", "", "Authentication key")
		zoneID     = flag.String("zone", "", "Zone ID")
		configFile = flag.String("file", "cloudflare_zone.json", "Config file")
		download   = flag.Bool("download", false, "Download config")
	)

	flag.Parse()

	if *download {
		settings := getSettings(*zoneID, *authEmail, *authKey)
		config := convertToConfig(settings)
		writeConfig(config, *configFile)
	} else {
		log.Fatalln("Save not implemented")
	}
}

func getSettings(zoneID, authEmail, authKey string) []ResponseSetting {
	url := fmt.Sprintf("%s/zones/%s/settings", rootURL, zoneID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalln("Constructing request failed:", err)
	}

	req.Header.Set("X-Auth-Email", authEmail)
	req.Header.Set("X-Auth-Key", authKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatalln("Request failed:", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalln("Incorrect HTTP response code:", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln("Reading response body failed:", err)
	}

	var parsedResp Response
	err = json.Unmarshal(body, &parsedResp)
	if err != nil {
		log.Fatalln("Parsing response body as JSON failed", err)
	}

	if !parsedResp.Success || len(parsedResp.Errors) > 0 {
		log.Fatalln("Response body indicated that request failed:", parsedResp)
	}

	return parsedResp.Result
}

func convertToConfig(settings []ResponseSetting) Config {
	config := make(Config)
	for _, setting := range settings {
		config[setting.ID] = setting.Value
	}

	return config
}

func writeConfig(config Config, file string) {
	bs, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		log.Fatalln("Parsing config to JSON failed:", err)
	}

	err = ioutil.WriteFile(file, bs, 0644)
	if err != nil {
		log.Fatalln("Writing config file failed:", err)
	}
}
