package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
)

type Profile struct {
	Project        string `json:"project"`
	ConnectionName string `json:"connection_name"`
	Instance       string `json:"instance"`
	Database       string `json:"database"`
	Port           int    `json:"port"`
}

func profilesPath() string {
	return filepath.Join(os.Getenv("HOME"), ".cloudsql", "profiles.json")
}

func loadProfiles() map[string]Profile {
	profiles := make(map[string]Profile)
	data, err := os.ReadFile(profilesPath())
	if err != nil {
		return profiles
	}
	json.Unmarshal(data, &profiles)
	return profiles
}

func saveProfiles(profiles map[string]Profile) {
	data, err := json.MarshalIndent(profiles, "", "  ")
	if err != nil {
		log.Fatalf("Error saving profiles: %v", err)
	}
	dir := filepath.Dir(profilesPath())
	os.MkdirAll(dir, 0755)
	if err := os.WriteFile(profilesPath(), data, 0644); err != nil {
		log.Fatalf("Error writing profiles: %v", err)
	}
}

func configSave(name string, port int, noConfig bool) {
	project := setProject(noConfig)
	connectionName := getInstance(project)
	parts := strings.Split(connectionName, ":")
	instanceName := parts[2]
	database := getDatabase(instanceName, project)

	profile := Profile{
		Project:        project,
		ConnectionName: connectionName,
		Instance:       instanceName,
		Database:       database,
		Port:           port,
	}

	profiles := loadProfiles()
	profiles[name] = profile
	saveProfiles(profiles)

	fmt.Printf("Profile %q saved\n", name)
}

func configList() {
	profiles := loadProfiles()
	if len(profiles) == 0 {
		fmt.Println("No profiles saved")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tPROJECT\tINSTANCE\tDATABASE\tPORT")
	for name, p := range profiles {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\n", name, p.Project, p.Instance, p.Database, p.Port)
	}
	w.Flush()
}

func configDelete(name string) {
	profiles := loadProfiles()
	if _, ok := profiles[name]; !ok {
		fmt.Printf("Profile %q not found\n", name)
		os.Exit(1)
	}
	delete(profiles, name)
	saveProfiles(profiles)
	fmt.Printf("Profile %q deleted\n", name)
}

func getProfile(name string) Profile {
	profiles := loadProfiles()
	p, ok := profiles[name]
	if !ok {
		fmt.Printf("Profile %q not found\n", name)
		os.Exit(1)
	}
	return p
}
