package cmd

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var rootCmd *cobra.Command
var connectCmd *cobra.Command
var disconnectCmd *cobra.Command
var listCmd *cobra.Command
var versionCmd *cobra.Command
var doctorCmd *cobra.Command
var infoCmd *cobra.Command

func init() {
	rootCmd = &cobra.Command{
		Use:   "cloudsql [sub]",
		Short: "CloudSQL CLI",
	}

	connectCmd = &cobra.Command{
		Use:   "connect [profile]",
		Short: "connect to cloudsql instance",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			port, err := cmd.Flags().GetInt("port")
			if err != nil {
				log.Fatalf("Error getting port: %v", err)
			}

			debug, err := cmd.Flags().GetBool("debug")
			if err != nil {
				log.Fatalf("Error getting debug: %v", err)
			}
			direct, err := cmd.Flags().GetBool("direct")
			if err != nil {
				log.Fatalf("Error getting direct: %v", err)
			}
			noIam, err := cmd.Flags().GetBool("no-iam")
			if err != nil {
				log.Fatalf("Error getting no-iam: %v", err)
			}
			timeout, err := cmd.Flags().GetDuration("timeout")
			if err != nil {
				log.Fatalf("Error getting timeout: %v", err)
			}
			useUnixSocket, err := cmd.Flags().GetBool("unix-socket")
			if err != nil {
				log.Fatalf("Error getting unix-socket: %v", err)
			}
			unixSocket, err := cmd.Flags().GetString("unix-socket-dir")
			if err != nil {
				log.Fatalf("Error getting unix-socket-dir: %v", err)
			}
			if useUnixSocket && unixSocket == "" {
				unixSocket = defaultUnixSocketDir()
			}

			profileName, _ := cmd.Flags().GetString("profile")
			if profileName == "" && len(args) > 0 {
				profileName = args[0]
			}

			if profileName != "" {
				p := getProfile(profileName)
				if port == 5432 {
					port = p.Port
				}
				connectInstanceWithProfile(p, port, debug, direct, noIam, timeout, unixSocket)
			} else {
				noConfig, err := cmd.Flags().GetBool("no-config")
				if err != nil {
					log.Fatalf("Error getting no-config: %v", err)
				}
				connectInstance(port, noConfig, debug, direct, noIam, timeout, unixSocket)
			}
		},
	}

	disconnectCmd = &cobra.Command{
		Use:   "disconnect",
		Short: "disconnect cloudsql instance",
		Run: func(cmd *cobra.Command, args []string) {
			all, _ := cmd.Flags().GetBool("all")
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
			fmt.Println("cloudsql 2.2.2")
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

	infoCmd = &cobra.Command{
		Use:   "info",
		Short: "show cloudsql instance details",
		Run: func(cmd *cobra.Command, args []string) {
			profileName, _ := cmd.Flags().GetString("profile")
			if profileName != "" {
				p := getProfile(profileName)
				showInstanceInfo(p.Project, p.Instance)
			} else {
				noConfig, err := cmd.Flags().GetBool("no-config")
				if err != nil {
					log.Fatalf("Error getting no-config: %v", err)
				}
				project := setProject(noConfig)
				connectionName := getInstance(project)
				parts := strings.Split(connectionName, ":")
				showInstanceInfo(project, parts[2])
			}
		},
	}

	configCmd := &cobra.Command{
		Use:   "config",
		Short: "manage connection profiles",
	}
	configSaveCmd := &cobra.Command{
		Use:   "save [name]",
		Short: "save a connection profile",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			port, _ := cmd.Flags().GetInt("port")
			noConfig, _ := cmd.Flags().GetBool("no-config")
			configSave(args[0], port, noConfig)
		},
	}
	configListCmd := &cobra.Command{
		Use:   "list",
		Short: "list saved profiles",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			configList()
		},
	}
	configDeleteCmd := &cobra.Command{
		Use:   "delete [name]",
		Short: "delete a saved profile",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			configDelete(args[0])
		},
	}
	configSaveCmd.Flags().Int("port", 5432, "port")
	configSaveCmd.Flags().BoolP("no-config", "", false, "load config from gcloud")
	configCmd.AddCommand(configSaveCmd, configListCmd, configDeleteCmd)

	rootCmd.AddCommand(disconnectCmd, connectCmd, listCmd, versionCmd, doctorCmd, completionCmd, infoCmd, configCmd, superviseCmd)
	infoCmd.Flags().BoolP("no-config", "", false, "load config from gcloud")
	infoCmd.Flags().StringP("profile", "p", "", "use saved profile")
	checkVersionCloudSqlProxy()
	connectCmd.PersistentFlags().Int("port", 5432, "port")
	connectCmd.Flags().BoolP("no-config", "", false, "load config from gcloud")
	connectCmd.Flags().BoolP("debug", "", false, "for troubleshooting. you can get cloud-sql-proxy log")
	connectCmd.Flags().BoolP("direct", "", false, "connect to cloudsql instance directly")
	connectCmd.Flags().BoolP("no-iam", "", false, "disable auto IAM authentication for cloud-sql-proxy")
	connectCmd.Flags().BoolP("unix-socket", "u", false, "use unix socket instead of TCP port (socket dir: ~/.cloudsql/sockets)")
	connectCmd.Flags().StringP("unix-socket-dir", "", "", "unix socket directory (implies --unix-socket)")
	connectCmd.Flags().StringP("profile", "p", "", "use saved profile")
	connectCmd.Flags().Duration("timeout", time.Hour, "automatically disconnect after this duration (0 to disable)")
	superviseCmd.Flags().Duration("ttl", time.Hour, "kill the proxy after this duration")
	disconnectCmd.Flags().BoolP("all", "a", false, "disconnect all cloudsql instance")
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatalf("Error executing command: %v", err)
	}
}
