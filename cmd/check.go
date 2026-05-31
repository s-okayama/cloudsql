package cmd

import (
	"fmt"
	"net"
	"os"
	"os/exec"

	"github.com/fatih/color"
)

func checkVersionCloudSqlProxy() {
	cloudsqlproxyversion := exec.Command("cloud-sql-proxy", "--version")
	_, err := cloudsqlproxyversion.Output()

	if err != nil {
		boldRed := color.New(color.FgRed, color.Bold)
		boldGreen := color.New(color.FgGreen, color.Bold)
		_, _ = boldRed.Printf("Error: %s\n", err)
		_, _ = boldRed.Println("Please upgrade your cloud-sql-proxy version to 2 or higher")
		_, _ = boldGreen.Println("Install URL:https://cloud.google.com/sql/docs/postgres/sql-proxy?hl=ja#install")
		os.Exit(0)
	}
}

func findAvailablePort(port int) int {
	ln, err := net.Listen("tcp4", fmt.Sprintf("127.0.0.1:%d", port))
	if err == nil {
		ln.Close()
		return port
	}

	yellow := color.New(color.FgYellow)
	boldYellow := yellow.Add(color.Bold)
	_, _ = boldYellow.Printf("Port %d is in use, searching for available port...\n", port)

	for p := port + 1; p <= port+100; p++ {
		ln, err = net.Listen("tcp4", fmt.Sprintf("127.0.0.1:%d", p))
		if err == nil {
			ln.Close()
			_, _ = boldYellow.Printf("Using port %d instead\n", p)
			return p
		}
	}

	fmt.Printf("No available port found in range %d-%d\n", port, port+100)
	os.Exit(1)
	return 0
}
