package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func listInstance() []string {
	command := fmt.Sprintf("ps aux  | grep cloud_sql_proxy | grep -v grep | awk -F '-instances=' '{print $NF}'")
	processlist := exec.Command("bash", "-c", command)
	output, _ := processlist.Output()
	line := strings.TrimSuffix(string(output), "\n")
	list := strings.Split(line, "\n")
	if list[0] == "" {
		fmt.Println("No Instance connected")
		os.Exit(1)
	}

	return list
}
