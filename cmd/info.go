package cmd

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/fatih/color"
	"google.golang.org/api/sqladmin/v1"
)

func showInstanceInfo(project, instance string) {
	ctx := context.Background()
	svc, err := sqladmin.NewService(ctx)
	if err != nil {
		log.Fatalf("Error creating sqladmin service: %v", err)
	}

	inst, err := svc.Instances.Get(project, instance).Context(ctx).Do()
	if err != nil {
		log.Fatalf("Error getting instance: %v", err)
	}

	bold := color.New(color.Bold)

	bold.Println("Instance Information")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("%-16s %s\n", "Instance:", inst.Name)
	fmt.Printf("%-16s %s\n", "Connection:", inst.ConnectionName)
	fmt.Printf("%-16s %s\n", "Project:", inst.Project)
	fmt.Printf("%-16s %s\n", "Region:", inst.Region)
	fmt.Printf("%-16s %s\n", "DB Version:", inst.DatabaseVersion)
	fmt.Printf("%-16s %s\n", "State:", inst.State)

	if inst.Settings != nil {
		fmt.Printf("%-16s %s\n", "Tier:", inst.Settings.Tier)
		fmt.Printf("%-16s %s\n", "Availability:", inst.Settings.AvailabilityType)

		if inst.Settings.DataDiskSizeGb > 0 {
			fmt.Printf("%-16s %d GB (%s)\n", "Storage:", inst.Settings.DataDiskSizeGb, inst.Settings.DataDiskType)
		}

		if inst.Settings.BackupConfiguration != nil && inst.Settings.BackupConfiguration.Enabled {
			fmt.Printf("%-16s Enabled (%s)\n", "Backup:", inst.Settings.BackupConfiguration.StartTime)
		} else {
			fmt.Printf("%-16s Disabled\n", "Backup:")
		}

		if inst.Settings.MaintenanceWindow != nil {
			days := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
			day := days[inst.Settings.MaintenanceWindow.Day]
			fmt.Printf("%-16s %s %02d:00\n", "Maintenance:", day, inst.Settings.MaintenanceWindow.Hour)
		}
	}

	privateIP := "(none)"
	publicIP := "(none)"
	for _, ip := range inst.IpAddresses {
		switch ip.Type {
		case "PRIVATE":
			privateIP = ip.IpAddress
		case "PRIMARY":
			publicIP = ip.IpAddress
		}
	}
	fmt.Printf("%-16s %s\n", "Private IP:", privateIP)
	fmt.Printf("%-16s %s\n", "Public IP:", publicIP)
	fmt.Printf("%-16s %s\n", "Created:", inst.CreateTime)
}
