package cmd

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cloud-sql-proxy-v2-operator [sub]",
	Short: "CloudSQL Proxy Operator",
}

var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "connect to cloudsql instance",
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetInt("port")
		_, err := net.Listen("tcp", ":"+strconv.Itoa(port))
		if err != nil {
			fmt.Printf("Port already in use\n")
			os.Exit(1)
		}
		connectInstance(port)
	},
}

var disconnectCmd = &cobra.Command{
	Use:   "disconnect",
	Short: "disconnect cloudsql instance",
	Run: func(cmd *cobra.Command, args []string) {
		disconnectInstance()
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list connected cloudsql instance",
	Run: func(cmd *cobra.Command, args []string) {
		for _, value := range listInstance() {
			fmt.Println(value)
		}
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of cloudsql",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("cloud-sql-proxy-v2-operator 0.1.4")
	},
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "troubleshooting",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		doctor()
	},
}

func Execute() {
	err := connectCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(disconnectCmd, connectCmd, listCmd, versionCmd, doctorCmd)
	connectCmd.PersistentFlags().Int("port", 5432, "port")
}
