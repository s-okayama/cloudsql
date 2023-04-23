package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func checkVersionCloudSqlProxy() {
	red := color.New(color.FgRed)
	boldRed := red.Add(color.Bold)
	green := color.New(color.FgGreen)
	boldGreen := green.Add(color.Bold)

	// Check Version cloud-sql-proxy
	cloudsqlproxyversion := exec.Command("cloud-sql-proxy", "--version")
	_, err := cloudsqlproxyversion.Output()

	if err != nil {
		_, _ = boldRed.Println("Error: %s", err)
		_, _ = boldRed.Println("Please upgrade your cloud-sql-proxy version to 2 or higher")
		_, _ = boldGreen.Println("Install URL:https://cloud.google.com/sql/docs/postgres/sql-proxy?hl=ja#install")
		os.Exit(0)
	}
}

func checkPort(port int) {
	red := color.New(color.FgRed)
	boldRed := red.Add(color.Bold)
	green := color.New(color.FgGreen)
	boldGreen := green.Add(color.Bold)

	//command := fmt.Sprintf("lsof -i tcp:%s", port)
	command := fmt.Sprintf("lsof -i tcp:" + strconv.Itoa(port))
	processlist := exec.Command("bash", "-c", command)
	output, _ := processlist.Output()
	line := strings.TrimSuffix(string(output), "\n")
	list := strings.Split(line, "\n")
	if list[0] != "" {
		_, _ = boldRed.Printf("Port \"%d\" already in use\n", port)
		_, _ = boldRed.Printf(string(output))
		_, _ = boldGreen.Printf("Can connect using:\n")
		_, _ = boldGreen.Printf("cloudsql connect --port 12345\n")
		os.Exit(0)
	}
}
