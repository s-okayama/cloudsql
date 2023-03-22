package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fatih/color"
)

func doctor() {
	red := color.New(color.FgRed)
	boldRed := red.Add(color.Bold)
	green := color.New(color.FgGreen)
	boldGreen := green.Add(color.Bold)

	// Check gcloud sdk
	gcloudversioncommand := fmt.Sprintf("gcloud version | head -n 1")
	gcloudversion := exec.Command("bash", "-c", gcloudversioncommand)
	gcloudversionout, err := gcloudversion.Output()
	checkErr := true
	if err != nil {
		_, _ = boldRed.Println("Please check gcloud sdk")
		_, _ = boldRed.Println("Error: %s", err)
		checkErr = false
	} else {
		fmt.Printf("gcloud version: %s", gcloudversionout)
	}

	// Check user is authenticated in gcloud
	gcloudusercommand := fmt.Sprintf("gcloud auth list --filter=status:ACTIVE --format='value(account)'")
	gclouduser := exec.Command("bash", "-c", gcloudusercommand)
	gclouduserout, err := gclouduser.Output()
	if err != nil {
		_, _ = boldRed.Println("User not authenticatedRun: gcloud auth application-default login")
		_, _ = boldRed.Println("Error: %s", err)
		checkErr = false
	} else {
		fmt.Printf("Authenticated user account: %s", gclouduserout)
	}

	// Check cloud-sql-proxy
	cloudsqlproxyversion := exec.Command("cloud-sql-proxy", "--version")
	cloudsqlproxyversionout, err := cloudsqlproxyversion.Output()
	if err != nil {
		_, _ = boldRed.Println("Please check cloud-sql-proxy")
		_, _ = boldRed.Println("Error: %s", err)
		checkErr = false
	} else {
		fmt.Printf("cloud-sql-proxy version: %s", cloudsqlproxyversionout)
	}

	// Check psql
	psqlversion := exec.Command("psql", "--version")
	psqlversionout, err := psqlversion.Output()
	if err != nil {
		_, _ = boldRed.Println("Please check psql")
		_, _ = boldRed.Println("Error: %s", err)
		checkErr = false
	} else {
		fmt.Printf("psql version: %s", psqlversionout)
	}

	// Check mysql
	mysqlversion := exec.Command("mysql", "--version")
	mysqlversionout, err := mysqlversion.Output()
	if err != nil {
		_, _ = boldRed.Println("Please check mysql")
		_, _ = boldRed.Println("Error: %s", err)
		checkErr = false
	} else {
		fmt.Printf("mysql version: %s", mysqlversionout)
	}

	// Check config file
	_, err = os.Stat(filepath.Join(os.Getenv("HOME"), "/.cloudsql/config"))
	if err != nil {
		_, _ = boldRed.Println("Please check config file ~/.cloudsql/config")
		_, _ = boldRed.Println("Error: %s", err)
		checkErr = false
	} else {
		fmt.Println("config file: ok")
	}

	if checkErr == true {
		_, _ = boldGreen.Println("Your system is All Green!")
	}
}

func checkVersionCloudSqlProxy() {
	red := color.New(color.FgRed)
	boldRed := red.Add(color.Bold)
	green := color.New(color.FgGreen)

	boldGreen := green.Add(color.Bold)

	// Check Version cloud-sql-proxy
	cloudsqlproxyversion := exec.Command("cloud-sql-proxy", "--version")
	cloudsqlproxyversionout, err := cloudsqlproxyversion.Output()
	if err != nil {
		_, _ = boldRed.Println("Error: %s", err)
		_, _ = boldRed.Println("Please upgrade your cloud-sql-proxy version to 2 or higher")
		_, _ = boldGreen.Println("Install URL:https://cloud.google.com/sql/docs/postgres/sql-proxy?hl=ja#install")
		os.Exit(1)

	} else {
		fmt.Printf("cloud-sql-proxy version: %s", cloudsqlproxyversionout)
	}
}
