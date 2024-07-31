package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/fatih/color"
)

func doctor() {
	red := color.New(color.FgRed)
	boldRed := red.Add(color.Bold)
	green := color.New(color.FgGreen)
	boldGreen := green.Add(color.Bold)

	// Find out the postgresSupportVersion from the following link.
	// https://github.com/google-cloud-sdk-unofficial/google-cloud-sdk/blame/7ffc79beeaa5ec5b900847691f3e047998033acf/lib/googlecloudsdk/generated_clients/apis/sqladmin/v1beta4/sqladmin_v1beta4_messages.py#L1175
	postgresSupportVersion := "486.0.0"

	// Check gcloud sdk
	gcloudversioncommand := fmt.Sprintf("gcloud version")
	gcloudversion := exec.Command("bash", "-c", gcloudversioncommand)
	gcloudversionout, err := gcloudversion.Output()

	re := regexp.MustCompile(`\d+\.\d+\.\d+`)
	version := re.FindString(string(gcloudversionout))
	fmt.Println("Google Cloud SDK Version:", version)

	checkErr := true
	if err != nil {
		_, _ = boldRed.Println("Please check gcloud sdk")
		_, _ = boldRed.Printf("Error: %s", err)
		checkErr = false
	} else if version < postgresSupportVersion {
		_, _ = boldRed.Printf("Your gcloud sdk version is %s. This version does not support Cloudsql 16.\n", version)
		_, _ = boldRed.Printf("You need to upgrade to at least above version %s.\n", postgresSupportVersion)
		checkErr = false
	} else {
		fmt.Printf("gcloud version: %s", version)
	}

	// Check user is authenticated in gcloud
	gcloudusercommand := fmt.Sprintf("gcloud auth list --filter=status:ACTIVE --format='value(account)'")
	gclouduser := exec.Command("bash", "-c", gcloudusercommand)
	gclouduserout, err := gclouduser.Output()
	if err != nil {
		_, _ = boldRed.Println("User not authenticatedRun: gcloud auth application-default login")
		_, _ = boldRed.Printf("Error: %s", err)
		checkErr = false
	} else {
		fmt.Printf("Authenticated user account: %s", gclouduserout)
	}

	// Check cloud-sql-proxy
	cloudsqlproxyversion := exec.Command("cloud-sql-proxy", "--version")
	cloudsqlproxyversionout, err := cloudsqlproxyversion.Output()
	if err != nil {
		_, _ = boldRed.Println("Please check cloud-sql-proxy")
		_, _ = boldRed.Printf("Error: %s", err)
		checkErr = false
	} else {
		fmt.Printf("cloud-sql-proxy version: %s", cloudsqlproxyversionout)
	}

	// Check psql
	psqlversion := exec.Command("psql", "--version")
	psqlversionout, err := psqlversion.Output()
	if err != nil {
		_, _ = boldRed.Println("Please check psql")
		_, _ = boldRed.Printf("Error: %s", err)
		checkErr = false
	} else {
		fmt.Printf("psql version: %s", psqlversionout)
	}

	// Check mysql
	mysqlversion := exec.Command("mysql", "--version")
	mysqlversionout, err := mysqlversion.Output()
	if err != nil {
		_, _ = boldRed.Println("Please check mysql")
		_, _ = boldRed.Printf("Error: %s", err)
		checkErr = false
	} else {
		fmt.Printf("mysql version: %s", mysqlversionout)
	}

	// Check config file
	_, err = os.Stat(filepath.Join(os.Getenv("HOME"), "/.cloudsql/config"))
	if err != nil {
		_, _ = boldRed.Println("Please check config file ~/.cloudsql/config")
		_, _ = boldRed.Printf("Error: %s", err)
		checkErr = false
	} else {
		fmt.Println("config file: ok")
	}

	if checkErr == true {
		_, _ = boldGreen.Println("Your system is All Green!")
	}
}
