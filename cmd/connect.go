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
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sqladmin/v1"
)

const (
	InfoColor   = "\033[1;34m%s\033[0m"
	NoticeColor = "\033[1;36m%s\033[0m"
	GreenColor  = "\033[32m"
)

func setProject() string {

	f, err := os.Open(filepath.Join(os.Getenv("HOME"), "/.cloudsql/config"))

	if err != nil {
		fmt.Println("error")
	}

	defer f.Close()

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
		Items:    []string(lines),
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

	c, err := google.DefaultClient(ctx, sqladmin.CloudPlatformScope)
	if err != nil {
		log.Fatal(err)
	}

	sqladminService, err := sqladmin.New(c)
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

func getInstance() (string, string) {
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

	_, result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("You choose %q\n", result)

	return result, project
}

func listdatabases(instance string, project string) []string {
	var list []string
	ctx := context.Background()

	c, err := google.DefaultClient(ctx, sqladmin.CloudPlatformScope)
	if err != nil {
		log.Fatal(err)
	}

	sqladminService, err := sqladmin.New(c)
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

func getDatabase(instance string, project string) string {
	databaseList := listdatabases(instance, project)

	searcher := func(input string, index int) bool {
		name := databaseList[index]
		input = strings.Replace(strings.ToLower(input), " ", "", -1)

		return strings.Contains(name, input)
	}

	prompt := promptui.Select{
		Label:    "Select Database",
		Items:    databaseList,
		Searcher: searcher,
		Stdout:   NoBellStdout,
	}

	_, result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
		return ""
	}

	return result
}

func connectInstance(port int) {
	var userName string
	var dbTypeName string
	var sqlInstanceName []string
	sqlConnectionName, project := getInstance()
	fmt.Println("Connecting Instance")
	sqlInstanceName = strings.Split(sqlConnectionName, ":")

	databaseList := getDatabase(sqlInstanceName[2], project)

	fmt.Println(databaseList)
	getdbtype := fmt.Sprintf("gcloud sql instances describe " + sqlInstanceName[2] + " --project=" + project + " --format='value(databaseVersion)'")

	dbtype := exec.Command("bash", "-c", getdbtype)
	getdbtypeOut, err1 := dbtype.Output()

	if err1 != nil {
		dbTypeName = "<dbtype>"
	} else {
		dbTypeName = strings.TrimSuffix(string(getdbtypeOut), "\n")
	}
	if strings.Contains(dbTypeName, "POSTGRES") {
		cmd := exec.Command("cloud_sql_proxy", "-enable_iam_login", "-instances="+sqlConnectionName+"=tcp:"+strconv.Itoa(port))
		cmd.Stdout = os.Stdout
		err := cmd.Start()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Cloudsql proxy process is running in background, process_id: %d\n", cmd.Process.Pid)

		command := fmt.Sprintf("gcloud auth list --filter=status:ACTIVE --format='value(account)'")
		user := exec.Command("bash", "-c", command)
		userOut, err := user.Output()
		if err != nil {
			userName = "<username>"
		} else {
			userName = strings.TrimSuffix(string(userOut), "\n")
		}

		color.Blue("%s", "Can connect using:")
		green := color.New(color.FgGreen)
		boldGreen := green.Add(color.Bold)
		boldGreen.Printf("psql -h localhost -U %s -p %d -d %s\n", userName, port, databaseList)
	}
	if strings.Contains(dbTypeName, "MYSQL") {
		cmd := exec.Command("cloud_sql_proxy", "-instances="+sqlConnectionName+"=tcp:3306")
		cmd.Stdout = os.Stdout
		err := cmd.Start()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Cloudsql proxy process is running in background, process_id: %d\n", cmd.Process.Pid)

		command := fmt.Sprintf("gcloud auth list --filter=status:ACTIVE --format='value(account)'")
		user := exec.Command("bash", "-c", command)
		userOut, err := user.Output()
		if err != nil {
			userName = "<username>"
		} else {
			userName = strings.TrimSuffix(string(userOut), "\n")
		}

		color.Blue("%s", "Can connect using:")
		green := color.New(color.FgGreen)
		var re = regexp.MustCompile("@.*")
		boldGreen := green.Add(color.Bold)
		boldGreen.Printf("mysql --user=%s --password=`gcloud auth print-access-token` --enable-cleartext-plugin --host=127.0.0.1 --port=3306 --database\n", re.ReplaceAllString(userName, ""))
	}
}
