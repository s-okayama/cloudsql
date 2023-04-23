package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "cloudsql [sub]",
	Short: "CloudSQL CLI",
}

var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "connect to cloudsql instance",
	Run: func(cmd *cobra.Command, args []string) {
		port, err := cmd.Flags().GetInt("port")
		if err != nil {
			os.Exit(0)
		}

		noConfig, err := cmd.Flags().GetBool("no-config")
		if err != nil {
			os.Exit(0)
		}
		checkPort(port)
		connectInstance(port, noConfig)
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
		fmt.Println("cloudsql 2.0.2")
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
		os.Exit(0)
	}
}

func init() {
	rootCmd.AddCommand(disconnectCmd, connectCmd, listCmd, versionCmd, doctorCmd)
	checkVersionCloudSqlProxy()
	connectCmd.PersistentFlags().Int("port", 5432, "port")
	connectCmd.Flags().BoolP("no-config", "", false, "load config from gcloud")
}
