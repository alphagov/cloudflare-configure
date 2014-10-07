package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
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

var (
	httpClient = &http.Client{}
	authEmail  = flag.String("email", "", "Authentication email address")
	authKey    = flag.String("key", "", "Authentication key")
	zoneID     = flag.String("zone", "", "Zone ID")
)

func main() {
	var (
		configFile = flag.String("file", "cloudflare_zone.json", "Config file")
		download   = flag.Bool("download", false, "Download configuration")
	)

	flag.Parse()

	settings := getSettings()
	config := convertToConfig(settings)

	if *download {
		log.Println("Saving configuration..")
		writeConfig(config, *configFile)
	} else {
		log.Println("Comparing and updating configuration..")
		configDesired := readConfig(*configFile)
		compareConfig(config, configDesired)
	}
}

func getSettings() []ResponseSetting {
	url := fmt.Sprintf("%s/zones/%s/settings", rootURL, *zoneID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalln("Constructing request failed:", err)
	}

	resp := makeRequest(req)
	return resp.Result
}

func makeRequest(req *http.Request) Response {
	req.Header.Set("X-Auth-Email", *authEmail)
	req.Header.Set("X-Auth-Key", *authKey)

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

	return parsedResp
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

func readConfig(file string) Config {
	bs, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalln("Reading config file failed:", err)
	}

	var config Config
	err = json.Unmarshal(bs, &config)
	if err != nil {
		log.Fatalln("Parsing config file as JSON failed:", err)
	}

	return config
}

func compareConfig(configActual, configDesired Config) {
	if reflect.DeepEqual(configActual, configDesired) {
		log.Println("No config changes to make")
		return
	}

	for key, valActual := range configActual {
		if valDesired, ok := configDesired[key]; ok {
			if reflect.DeepEqual(valActual, valDesired) {
				continue
			}

			log.Println("Change:", key, valActual, "->", valDesired)
		} else {
			log.Println("Not in config:", key)
		}
	}
}
