package main

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/jwaldrip/odin.v1/cli"
)

const flagRequiredDefault = "REQUIRED"

var app = cli.New(Version, "CloudFlare Configure", exitWithUsage)

func init() {
	app.DefineStringFlag("email", flagRequiredDefault, "Authentication email address")
	app.DefineStringFlag("key", flagRequiredDefault, "Authentication key")

	zones := app.DefineSubCommand("zones", "List available zones by name and ID", zones)
	zones.InheritFlags("email", "key")

	download := app.DefineSubCommand("download", "Download configuration to file", download)
	download.InheritFlags("email", "key")
	download.DefineParams("zone_id", "file")

	upload := app.DefineSubCommand("upload", "Upload configuration from file", upload)
	upload.InheritFlags("email", "key")
	upload.DefineParams("zone_id", "file")
	upload.DefineBoolFlag("dry-run", false, "Log changes without actioning them")
}

func main() {
	app.Start()
}

func setup(cmd cli.Command) *CloudFlare {
	query := &CloudFlareQuery{
		AuthEmail: getRequiredFlag(cmd, "email"),
		AuthKey:   getRequiredFlag(cmd, "key"),
		RootURL:   "https://api.cloudflare.com/v4",
	}
	logger := log.New(os.Stdout, "", log.LstdFlags)

	return NewCloudFlare(query, logger)
}

func getRequiredFlag(cmd cli.Command, name string) string {
	val := cmd.Flag(name).String()
	if val == flagRequiredDefault {
		fmt.Print("missing flag: ", name, "\n\n")
		exitWithUsage(cmd)
	}

	return val
}

func exitWithUsage(cmd cli.Command) {
	cmd.Usage()
	os.Exit(2)
}

func zones(cmd cli.Command) {
	cloudflare := setup(cmd)
	zones, err := cloudflare.Zones()
	if err != nil {
		log.Fatalln(err)
	}

	for _, zone := range zones {
		fmt.Println(zone.ID, "\t", zone.Name)
	}
}

func download(cmd cli.Command) {
	cloudflare := setup(cmd)
	settings, err := cloudflare.Settings(cmd.Param("zone_id").String())
	if err != nil {
		log.Fatalln(err)
	}

	file := cmd.Param("file").String()
	log.Println("Saving config to:", file)

	err = SaveConfigItems(settings.ConfigItems(), file)
	if err != nil {
		log.Fatalln(err)
	}
}

func upload(cmd cli.Command) {
	cloudflare := setup(cmd)
	zone := cmd.Param("zone_id").String()

	settings, err := cloudflare.Settings(zone)
	if err != nil {
		log.Fatalln(err)
	}

	configActual := settings.ConfigItems()
	configDesired, err := LoadConfigItems(cmd.Param("file").String())
	if err != nil {
		log.Fatalln(err)
	}

	configUpdate, err := CompareConfigItemsForUpdate(configActual, configDesired)
	if err != nil {
		log.Fatalln(err)
	}

	logOnly := (cmd.Flag("dry-run").Get() == true)
	err = cloudflare.Update(zone, configUpdate, logOnly)
	if err != nil {
		log.Fatalln(err)
	}
}
