package cmd

import (
	"bufio"
	"context"
	"fmt"
	"google.golang.org/api/sqladmin/v1"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
)

func setProject(noConfig bool) string {
	if noConfig {
		var projId []string
		getprojectcommand := fmt.Sprintf("gcloud projects list --format='value(project_id)'")
		getproject := exec.Command("bash", "-c", getprojectcommand)
		getprojectout, err := getproject.Output()
		if err != nil {
			log.Fatal(err)
		} else {
			proj := strings.TrimSuffix(string(getprojectout), "\n")
			projId = strings.Split(proj, "\n")
		}
		searcher := func(input string, index int) bool {
			name := projId[index]
			input = strings.Replace(strings.ToLower(input), " ", "", -1)
			return strings.Contains(name, input)
		}

		prompt := promptui.Select{
			Label:    "Select GCP Project",
			Items:    projId,
			Searcher: searcher,
			Stdout:   NoBellStdout,
		}

		_, result, err := prompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			os.Exit(1)
			return ""
		}

		fmt.Printf("Project ID: %q\n", result)
		promptresult := strings.Split(result, ":")
		projectId := promptresult[len(promptresult)-1]
		return projectId

	} else {
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

func getInstance(project string) string {
	instancelist := listInstances(project)

	searcher := func(input string, index int) bool {
		name := instancelist[index]
		input = strings.Replace(strings.ToLower(input), " ", "", -1)
		return strings.Contains(name, input)
	}

	prompt := promptui.Select{
		Label:    "Select Instance",
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

	return result
}

func listDatabases(instance string, project string) []string {
	var list []string
	ctx := context.Background()

	sqladminService, err := sqladmin.NewService(ctx)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := sqladminService.Databases.List(project, instance).Context(ctx).Do()
	if err != nil {
		log.Fatal(err)
	}

	if resp == nil || resp.Items == nil {
		log.Fatal("No databases found or unable to retrieve databases")
	}

	for _, database := range resp.Items {
		list = append(list, database.Name)
	}

	return list
}

func getDatabase(instance string, project string) string {
	databaseList := listDatabases(instance, project)

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
		os.Exit(0)
		return ""
	}

	return result
}

func getUser() string {
	var userName string
	command := fmt.Sprintf("gcloud auth list --filter=status:ACTIVE --format='value(account)'")
	user := exec.Command("bash", "-c", command)
	userOut, err := user.Output()
	if err != nil {
		userName = "<username>"
	} else {
		userName = strings.TrimSuffix(string(userOut), "\n")
	}
	return userName
}

func getdbTypeName(instance string, project string) string {
	var result string

	getdbtype := fmt.Sprintf("gcloud sql instances describe " + instance + " --project=" + project + " --format='value(databaseVersion)'")
	dbtype := exec.Command("bash", "-c", getdbtype)
	getdbtypeOut, err1 := dbtype.Output()

	if err1 != nil {
		result = "<dbtype>"
	} else {
		result = strings.TrimSuffix(string(getdbtypeOut), "\n")
	}
	if result == "" || result == "<dbtype>" {
		fmt.Println("Error : You do not have permissions to CloudSQL or there is a problem with the gcloud command")
		os.Exit(0)
	}
	return result
}

func connectInstance(port int, noConfig bool, debug bool) {
	var dbTypeName string
	var sqlInstanceName []string
	var sqlConnectionName string
	var userName string

	// color setting
	green := color.New(color.FgGreen)
	blue := color.New(color.FgBlue)
	boldGreen := green.Add(color.Bold)
	boldBlue := blue.Add(color.Bold)

	// select database
	project := setProject(noConfig)
	sqlConnectionName = getInstance(project)
	sqlInstanceName = strings.Split(sqlConnectionName, ":")
	databaseList := getDatabase(sqlInstanceName[2], project)
	dbTypeName = getdbTypeName(sqlInstanceName[2], project)
	userName = getUser()

	// connect instance
	if strings.Contains(dbTypeName, "POSTGRES") {
		if debug {
			command := fmt.Sprintf("cloud-sql-proxy %s --auto-iam-authn --debug --private-ip --port=%d", sqlConnectionName, port)
			color.Blue("[Debug Mode]\nThe following commands are executed in the background.\n")
			_, _ = boldBlue.Printf("%s\n", command)
			color.Green("Can connect using:\n")
			_, _ = boldGreen.Printf("psql -h localhost -U %s -p %d -d %s\n", userName, port, databaseList)
			debug := exec.Command("bash", "-c", command)
			debug.Stdout = os.Stdout
			debug.Stderr = os.Stderr
			err := debug.Run()
			if err != nil {
				log.Fatal(err)
			}

		} else {
			cmd := exec.Command("cloud-sql-proxy", sqlConnectionName, "--auto-iam-authn", "--private-ip", "--quiet", "--port="+strconv.Itoa(port))
			cmd.Stdout = os.Stdout
			err := cmd.Start()
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Cloudsql proxy process is running in background, process_id: %d\n", cmd.Process.Pid)

			color.Blue("Can connect using:")
			_, _ = boldGreen.Printf("psql -h localhost -U %s -p %d -d %s\n", userName, port, databaseList)
		}
	}
	if strings.Contains(dbTypeName, "MYSQL") {
		if debug {
			command := fmt.Sprintf("cloud-sql-proxy %s --auto-iam-authn --private-ip --debug --port=%d", sqlConnectionName, port)
			color.Blue("[Debug Mode]\nThe following commands are executed in the background.\n")
			_, _ = boldBlue.Printf("%s\n", command)
			color.Green("Can connect using:\n")
			var re = regexp.MustCompile("@.*")
			_, _ = boldGreen.Printf("mysql --user=%s --password=`gcloud auth print-access-token` --enable-cleartext-plugin --host=127.0.0.1 --port=%d\n", re.ReplaceAllString(userName, ""), port)
			debug := exec.Command("bash", "-c", command)
			debug.Stdout = os.Stdout
			debug.Stderr = os.Stderr
			err := debug.Run()
			if err != nil {
				log.Fatal(err)
			}
		}

		port = 3306
		cmd := exec.Command("cloud-sql-proxy", sqlConnectionName, "--auto-iam-authn", "--private-ip", "--quiet", "--port="+strconv.Itoa(port))
		cmd.Stdout = os.Stdout
		err := cmd.Start()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Cloudsql proxy process is running in background, process_id: %d\n", cmd.Process.Pid)

		color.Blue("Can connect using:")
		var re = regexp.MustCompile("@.*")
		_, _ = boldGreen.Printf("mysql --user=%s --password=`gcloud auth print-access-token` --enable-cleartext-plugin --host=127.0.0.1 --port=%d --database=%s\n", re.ReplaceAllString(userName, ""), port, databaseList)
		// Temporarily commented out when database is selected because the connection is not possible due to permission issues.
		//_, _ = boldGreen.Printf("mysql --user=%s --password=`gcloud auth print-access-token` --enable-cleartext-plugin --host=127.0.0.1 --port=%d --database=%s\n", re.ReplaceAllString(userName, ""), port, databaseList)
	}
}
