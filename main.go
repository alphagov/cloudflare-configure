package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
)

type ConfigItems map[string]interface{}

type CloudFlareResponse struct {
	Success  bool
	Errors   []string
	Messages []string
	Result   json.RawMessage
}

type CloudFlareZoneItem struct {
	ID   string
	Name string
}

type CloudFlareConfigItem struct {
	ID         string
	Value      interface{}
	ModifiedOn string `json:"modified_on"`
	Editable   bool
}

type CloudFlareRequestItem struct {
	Value interface{} `json:"value"`
}

const RootURL = "https://api.cloudflare.com/v4"

var (
	httpClient = &http.Client{}
	authEmail  = flag.String("email", "", "Authentication email address [required]")
	authKey    = flag.String("key", "", "Authentication key [required]")
	zoneID     = flag.String("zone", "", "Zone ID [required]")
)

func main() {
	var (
		configFile = flag.String("file", "", "Config file [required]")
		download   = flag.Bool("download", false, "Download configuration")
		listZones  = flag.Bool("list-zones", false, "List zone IDs and names")
		dryRun     = flag.Bool("dry-run", false, "Don't submit changes")
	)

	flag.Parse()
	checkRequiredFlags([]string{"email", "key"})

	if *listZones {
		zones := getZones()
		printZones(zones)
		return
	}

	checkRequiredFlags([]string{"zone", "file"})
	settings := getSettings()
	config := convertToConfig(settings)

	if *download {
		log.Println("Saving configuration..")
		writeConfig(config, *configFile)
		return
	}

	if *dryRun {
		log.Println("Dry run mode. Changes won't be submitted")
	}
	log.Println("Comparing and updating configuration..")
	configDesired := readConfig(*configFile)
	compareAndUpdate(config, configDesired, *dryRun)
}

// Ensure that all mandatory flags have been provided.
func checkRequiredFlags(names []string) {
	for _, name := range names {
		f := flag.Lookup(name)
		if f.Value.String() == f.DefValue {
			flag.Usage()
			os.Exit(1)
		}
	}
}

// Get all available zones.
func getZones() []CloudFlareZoneItem {
	url := fmt.Sprintf("%s/zones", RootURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalln("Constructing request failed:", err)
	}

	resp := makeRequest(req)

	var zones []CloudFlareZoneItem
	err = json.Unmarshal(resp.Result, &zones)
	if err != nil {
		log.Fatalln("Parsing results as JSON failed", err)
	}

	return zones
}

// Modify the value of a setting. Assumes that the name of the API endpoint
// matches the key.
func changeSetting(key string, value interface{}) {
	url := fmt.Sprintf("%s/zones/%s/settings/%s", RootURL, *zoneID, key)

	body, err := json.Marshal(CloudFlareRequestItem{value})
	if err != nil {
		log.Fatalln("Parsing request JSON failed:", err)
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(body))
	if err != nil {
		log.Fatalln("Constructing request failed:", err)
	}

	_ = makeRequest(req)
}

// Fetch all settings for a zone.
func getSettings() []CloudFlareConfigItem {
	url := fmt.Sprintf("%s/zones/%s/settings", RootURL, *zoneID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalln("Constructing request failed:", err)
	}

	resp := makeRequest(req)

	var settings []CloudFlareConfigItem
	err = json.Unmarshal(resp.Result, &settings)
	if err != nil {
		log.Fatalln("Parsing results as JSON failed", err)
	}

	return settings
}

// Add authentication headers to an API request, submit it, check for
// errors, and parse the response body as JSON.
func makeRequest(req *http.Request) CloudFlareResponse {
	req.Header.Set("X-Auth-Email", *authEmail)
	req.Header.Set("X-Auth-Key", *authKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatalln("Request failed:", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln("Reading response body failed:", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalln("Incorrect HTTP response code", resp.StatusCode, ":", string(body))
	}

	var parsedResp CloudFlareResponse
	err = json.Unmarshal(body, &parsedResp)
	if err != nil {
		log.Fatalln("Parsing response body as JSON failed", err)
	}

	if !parsedResp.Success || len(parsedResp.Errors) > 0 {
		log.Fatalln("Response body indicated that request failed:", parsedResp)
	}

	return parsedResp
}

// Output zone IDs and names.
func printZones(zones []CloudFlareZoneItem) {
	for _, zone := range zones {
		fmt.Println(zone.ID, "\t", zone.Name)
	}
}

// Convert an array-of-maps that represent config items into a flat map that
// is more human readable and easier to check for the existence of keys.
func convertToConfig(settings []CloudFlareConfigItem) ConfigItems {
	config := make(ConfigItems)
	for _, setting := range settings {
		config[setting.ID] = setting.Value
	}

	return config
}

func writeConfig(config ConfigItems, file string) {
	bs, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		log.Fatalln("Parsing config to JSON failed:", err)
	}

	err = ioutil.WriteFile(file, bs, 0644)
	if err != nil {
		log.Fatalln("Writing config file failed:", err)
	}
}

// Load a JSON config from disk.
func readConfig(file string) ConfigItems {
	bs, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalln("Reading config file failed:", err)
	}

	var config ConfigItems
	err = json.Unmarshal(bs, &config)
	if err != nil {
		log.Fatalln("Parsing config file as JSON failed:", err)
	}

	return config
}

// Compare two ConfigItems. Log a message if a key name appears in one but
// not the other. Submit changes if the actual values doesn't match desired.
func compareAndUpdate(configActual, configDesired ConfigItems, dryRun bool) {
	if reflect.DeepEqual(configActual, configDesired) {
		log.Println("No config changes to make")
		return
	}

	for key, val := range configDesired {
		if _, ok := configActual[key]; !ok {
			log.Println("Missing from remote config:", key, val)
		}
	}

	for key, valActual := range configActual {
		if valDesired, ok := configDesired[key]; !ok {
			log.Println("Missing from local config:", key, valActual)
		} else if !reflect.DeepEqual(valActual, valDesired) {
			log.Println("Changing setting:", key, valActual, "->", valDesired)
			if !dryRun {
				changeSetting(key, valDesired)
			}
		}
	}
}
