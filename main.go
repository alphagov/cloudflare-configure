package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

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
	logger := log.New(os.Stdout, "", log.LstdFlags)
	cloudflare := NewCloudFlare(query, logger)

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
	config := settings.ConfigItems()

	if *download {
		log.Println("Saving configuration..")
		writeConfig(config, *configFile)
		return
	}

	configDesired := readConfig(*configFile)
	configUpdate, err := CompareConfigItemsForUpdate(config, configDesired)
	if err != nil {
		log.Fatalln(err)
	}

	cloudflare.Update(*zoneID, configUpdate, *dryRun)
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
