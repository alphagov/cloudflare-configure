package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
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

func main() {
	var (
		authEmail  = flag.String("email", "", "Authentication email address [required]")
		authKey    = flag.String("key", "", "Authentication key [required]")
		zoneID     = flag.String("zone", "", "Zone ID [required]")
		configFile = flag.String("file", "", "Config file [required]")
		download   = flag.Bool("download", false, "Download configuration")
		listZones  = flag.Bool("list-zones", false, "List zone IDs and names")
		dryRun     = flag.Bool("dry-run", false, "Don't submit changes")
	)

	flag.Parse()
	checkRequiredFlags([]string{"email", "key"})

	query := &CloudFlareQuery{
		AuthEmail: *authEmail,
		AuthKey:   *authKey,
		RootURL:   RootURL,
	}
	cloudflare := NewCloudFlare(query)

	if *listZones {
		zones, err := cloudflare.Zones()
		if err != nil {
			log.Fatal("Couldn't get zones", err)
		}
		printZones(zones)
		return
	}

	checkRequiredFlags([]string{"zone", "file"})
	settings, err := cloudflare.Settings(*zoneID)
	if err != nil {
		log.Fatal("Couldn't read settings", err)
	}
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
	compareAndUpdate(cloudflare, *zoneID, config, configDesired, *dryRun)
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
func compareAndUpdate(cloudflare *CloudFlare, zoneID string, configActual, configDesired ConfigItems, dryRun bool) {
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
				cloudflare.Set(zoneID, key, valDesired)
			}
		}
	}
}
