package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestProfiles(t *testing.T) (string, func()) {
	t.Helper()
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	os.MkdirAll(filepath.Join(tmpDir, ".cloudsql"), 0755)
	return tmpDir, func() { os.Setenv("HOME", origHome) }
}

func TestLoadProfiles_Empty(t *testing.T) {
	_, cleanup := setupTestProfiles(t)
	defer cleanup()

	profiles := loadProfiles()
	if len(profiles) != 0 {
		t.Errorf("expected empty profiles, got %d", len(profiles))
	}
}

func TestSaveAndLoadProfiles(t *testing.T) {
	_, cleanup := setupTestProfiles(t)
	defer cleanup()

	profiles := map[string]Profile{
		"mydb": {
			Project:        "test-project",
			ConnectionName: "test-project:asia-northeast1:mydb",
			Instance:       "mydb",
			Database:       "app_production",
			Port:           5432,
		},
	}
	saveProfiles(profiles)

	loaded := loadProfiles()
	if len(loaded) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(loaded))
	}
	p := loaded["mydb"]
	if p.Project != "test-project" {
		t.Errorf("expected project test-project, got %s", p.Project)
	}
	if p.ConnectionName != "test-project:asia-northeast1:mydb" {
		t.Errorf("expected connection name test-project:asia-northeast1:mydb, got %s", p.ConnectionName)
	}
	if p.Database != "app_production" {
		t.Errorf("expected database app_production, got %s", p.Database)
	}
	if p.Port != 5432 {
		t.Errorf("expected port 5432, got %d", p.Port)
	}
}

func TestSaveMultipleProfiles(t *testing.T) {
	_, cleanup := setupTestProfiles(t)
	defer cleanup()

	profiles := map[string]Profile{
		"db1": {Project: "proj1", Instance: "inst1", Database: "db1", Port: 5432},
		"db2": {Project: "proj2", Instance: "inst2", Database: "db2", Port: 5433},
	}
	saveProfiles(profiles)

	loaded := loadProfiles()
	if len(loaded) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(loaded))
	}
}

func TestConfigDelete(t *testing.T) {
	_, cleanup := setupTestProfiles(t)
	defer cleanup()

	profiles := map[string]Profile{
		"mydb":  {Project: "proj1", Instance: "inst1", Database: "db1", Port: 5432},
		"other": {Project: "proj2", Instance: "inst2", Database: "db2", Port: 5433},
	}
	saveProfiles(profiles)

	configDelete("mydb")

	loaded := loadProfiles()
	if len(loaded) != 1 {
		t.Fatalf("expected 1 profile after delete, got %d", len(loaded))
	}
	if _, ok := loaded["mydb"]; ok {
		t.Error("expected mydb to be deleted")
	}
	if _, ok := loaded["other"]; !ok {
		t.Error("expected other to still exist")
	}
}

func TestProfilesPath(t *testing.T) {
	path := profilesPath()
	if !filepath.IsAbs(path) {
		t.Errorf("expected absolute path, got %s", path)
	}
	if filepath.Base(path) != "profiles.json" {
		t.Errorf("expected profiles.json, got %s", filepath.Base(path))
	}
}
