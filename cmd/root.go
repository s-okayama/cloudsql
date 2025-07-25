package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
)

var rootCmd *cobra.Command
var connectCmd *cobra.Command
var disconnectCmd *cobra.Command
var listCmd *cobra.Command
var versionCmd *cobra.Command
var doctorCmd *cobra.Command

func init() {
	rootCmd = &cobra.Command{
		Use:   "cloudsql [sub]",
		Short: "CloudSQL CLI",
	}

	connectCmd = &cobra.Command{
		Use:   "connect",
		Short: "connect to cloudsql instance",
		Run: func(cmd *cobra.Command, args []string) {
			port, err := cmd.Flags().GetInt("port")
			if err != nil {
				log.Fatalf("Error getting port: %v", err)
			}

			noConfig, err := cmd.Flags().GetBool("no-config")
			if err != nil {
				log.Fatalf("Error getting no-config: %v", err)
			}

			debug, err := cmd.Flags().GetBool("debug")
			if err != nil {
				log.Fatalf("Error getting debug: %v", err)
			}
			direct, err := cmd.Flags().GetBool("direct")
			if err != nil {
				log.Fatalf("Error getting direct: %v", err)
			}
			checkPort(port)
			connectInstance(port, noConfig, debug, direct)
		},
	}

	disconnectCmd = &cobra.Command{
		Use:   "disconnect",
		Short: "disconnect cloudsql instance",
		Run: func(cmd *cobra.Command, args []string) {
			all, err := cmd.Flags().GetBool("all")
			if err != nil {
				log.Fatalf("Error getting all: %v", err)
			}
			disconnectInstance(all)
		},
	}

	listCmd = &cobra.Command{
		Use:   "list",
		Short: "list connected cloudsql instance",
		Run: func(cmd *cobra.Command, args []string) {
			for _, value := range listInstance() {
				fmt.Println(value)
			}
		},
	}

	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number of cloudsql",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("cloudsql 2.1.0")
		},
	}

	doctorCmd = &cobra.Command{
		Use:   "doctor",
		Short: "troubleshooting",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			doctor()
		},
	}

	rootCmd.AddCommand(disconnectCmd, connectCmd, listCmd, versionCmd, doctorCmd)
	checkVersionCloudSqlProxy()
	connectCmd.PersistentFlags().Int("port", 5432, "port")
	connectCmd.Flags().BoolP("no-config", "", false, "load config from gcloud")
	connectCmd.Flags().BoolP("debug", "", false, "for troubleshooting. you can get cloud-sql-proxy log")
	connectCmd.Flags().BoolP("direct", "", false, "connect to cloudsql instance directly")
	disconnectCmd.Flags().BoolP("all", "a", false, "disconnect all cloudsql instance")
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatalf("Error executing command: %v", err)
	}
}
