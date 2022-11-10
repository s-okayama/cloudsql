package cmd

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cloudsql [sub]",
	Short: "CloudSQL CLI",
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

func Execute() {
	err := connectCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(disconnectCmd, connectCmd)
	connectCmd.PersistentFlags().Int("port", 5432, "port")
}
