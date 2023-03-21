package cmd

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"google.golang.org/api/sqladmin/v1"
)

func setProject() string {
	f, err := os.Open(filepath.Join(os.Getenv("HOME"), "/.cloudsql/config"))

	if err != nil {
		fmt.Println("error")
	}

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			fmt.Println("error")
		}
	}(f)

	lines := make([]string, 0, 100)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	searcher := func(input string, index int) bool {
		name := lines[index]
		input = strings.Replace(strings.ToLower(input), " ", "", -1)

		return strings.Contains(name, input)
	}

	prompt := promptui.Select{
		Label:    "Select Project",
		Items:    lines,
		Searcher: searcher,
		Stdout:   NoBellStdout,
	}

	_, result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
		return ""
	}

	fmt.Printf("You choose %q\n", result)
	return result
}

func listInstances(project string) []string {
	var list []string
	ctx := context.Background()

	sqladminService, err := sqladmin.NewService(ctx)
	if err != nil {
		log.Fatal(err)
	}

	req := sqladminService.Instances.List(project)
	if err := req.Pages(ctx, func(page *sqladmin.InstancesListResponse) error {
		for _, databaseInstance := range page.Items {
			list = append(list, databaseInstance.ConnectionName)
		}
		return nil
	}); err != nil {
		log.Fatal(err)
	}
	return list
}

func (c *Config) getInstance() {
	project := setProject()
	instancelist := listInstances(project)

	searcher := func(input string, index int) bool {
		name := instancelist[index]
		input = strings.Replace(strings.ToLower(input), " ", "", -1)

		return strings.Contains(name, input)
	}

	prompt := promptui.Select{
		Label:    "Select Project" + project,
		Items:    instancelist,
		Searcher: searcher,
		Stdout:   NoBellStdout,
	}

	_, sqlConnectionName, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("You choose %q\n", sqlConnectionName)
	c.SetProject(project)
	c.SetSqlConnectionName(sqlConnectionName)
}

func listDatabases(instance string, project string) []string {
	var list []string
	ctx := context.Background()

	sqladminService, err := sqladmin.NewService(ctx)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := sqladminService.Databases.List(project, instance).Context(ctx).Do()
	for _, database := range resp.Items {
		list = append(list, database.Name)
	}
	if err != nil {
		log.Fatal(err)
	}

	return list
}

func (c *Config) getDb() {
	c.dbList = listDatabases(c.sqlInstanceName, c.project)

	searcher := func(input string, index int) bool {
		name := c.dbList[index]
		input = strings.Replace(strings.ToLower(input), " ", "", -1)

		return strings.Contains(name, input)
	}

	prompt := promptui.Select{
		Label:    "Select Database",
		Items:    c.dbList,
		Searcher: searcher,
		Stdout:   NoBellStdout,
	}

	_, dbName, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}
	c.SetDbName(dbName)
}

func (c *Config) getDbType() {
	getDbTypeCommand := fmt.Sprintf("gcloud sql instances describe " + c.sqlInstanceName + " --project=" + c.project + " --format='value(databaseVersion)'")
	execDbTypeCommand := exec.Command("bash", "-c", getDbTypeCommand)
	dbType, err := execDbTypeCommand.Output()

	if err != nil {
		c.SetDbType("<dbtype>")
	} else {
		c.SetDbType(strings.TrimSuffix(string(dbType), "\n"))
	}
}

func connectWithPrivateIp() bool {
	prompt := promptui.Select{
		Label:  "Types of ip address",
		Items:  []string{"Public ip", "Private ip"},
		Stdout: NoBellStdout,
	}

	_, result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return false
	}
	fmt.Printf("You choose %q\n", result)

	return result == "Private ip"
}

func (c *Config) connectInstanceWithCloudSqlProxy() {
	bin := "cloud-sql-proxy "
	options := "--auto-iam-authn --address 0.0.0.0 --port " + strconv.Itoa(c.port) + " "

	if connectWithPrivateIp() {
		options += "--private-ip "
	}
	cmdstr := bin + options + c.sqlConnectionName

	cmd := exec.Command("bash", "-c", cmdstr)
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Cloudsql proxy process is running in background, process_id: %d\n", cmd.Process.Pid)
}

func (c *Config) getUserName() {
	command := "gcloud auth list --filter=status:ACTIVE --format='value(account)'"
	user := exec.Command("bash", "-c", command)
	userOut, err := user.Output()
	if err != nil {
		c.SetUserName("<username>")
	} else {
		c.SetUserName(strings.TrimSuffix(string(userOut), "\n"))
	}
}

func (c *Config) showConnectionMethod() {
	color.Blue("%s", "Can connect using:")
	green := color.New(color.FgGreen)
	boldGreen := green.Add(color.Bold)
	if strings.Contains(c.dbType, "POSTGRES") {
		_, _ = boldGreen.Printf("psql -h localhost -U %s -p %d -d %s\n", c.userName, c.port, c.dbList)
	}
	if strings.Contains(c.dbType, "MYSQL") {
		var re = regexp.MustCompile("@.*")
		_, _ = boldGreen.Printf("mysql --user=%s --password=`gcloud auth print-access-token` --enable-cleartext-plugin --host=127.0.0.1 --port=%d --database=%s\n", re.ReplaceAllString(c.userName, ""), c.port, c.dbList)
	}
}

func (c *Config) connectInstance() {
	c.getInstance()
	fmt.Println("Connecting Instance")
	c.getDb()
	fmt.Println(c.dbName)
	c.getDbType()
	c.connectInstanceWithCloudSqlProxy()
	c.getUserName()
	c.showConnectionMethod()
}
