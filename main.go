package main

import (
	"flag"
	"fmt"
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
		err := SaveConfigItems(config, *configFile)
		if err != nil {
			log.Fatalln(err)
		}

		return
	}

	configDesired, err := LoadConfigItems(*configFile)
	if err != nil {
		log.Fatalln(err)
	}

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
