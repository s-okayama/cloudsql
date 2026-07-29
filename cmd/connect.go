package cmd

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"google.golang.org/api/sqladmin/v1"

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
	database := listDatabases(instance, project)

	searcher := func(input string, index int) bool {
		name := database[index]
		input = strings.Replace(strings.ToLower(input), " ", "", -1)
		return strings.Contains(name, input)
	}

	prompt := promptui.Select{
		Label:    "Select Database",
		Items:    database,
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

	getdbtype := fmt.Sprintf("gcloud sql instances describe %s --project=%s --format='value(databaseVersion)'", instance, project)
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

func defaultUnixSocketDir() string {
	return filepath.Join(os.Getenv("HOME"), ".cloudsql", "sockets")
}

func findExistingProxyProcess(connectionName string) bool {
	command := fmt.Sprintf("ps aux | grep cloud-sql-proxy | grep -v grep | grep '%s'", connectionName)
	out, err := exec.Command("bash", "-c", command).Output()
	return err == nil && len(out) > 0
}

func findExistingProxy(connectionName string) int {
	command := fmt.Sprintf("ps aux | grep cloud-sql-proxy | grep -v grep | grep '%s'", connectionName)
	out, err := exec.Command("bash", "-c", command).Output()
	if err != nil || len(out) == 0 {
		return 0
	}
	re := regexp.MustCompile(`--port[= ](\d+)`)
	match := re.FindStringSubmatch(string(out))
	if len(match) < 2 {
		return 0
	}
	port, _ := strconv.Atoi(match[1])

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 1*time.Second)
	if err != nil {
		return 0
	}
	conn.Close()
	return port
}

func connectInstanceWithProfile(p Profile, port int, debug bool, direct bool, noIam bool, timeout time.Duration, unixSocket string) {
	doConnect(p.Project, p.ConnectionName, p.Instance, p.Database, port, debug, direct, noIam, timeout, unixSocket)
}

func connectInstance(port int, noConfig bool, debug bool, direct bool, noIam bool, timeout time.Duration, unixSocket string) {
	project := setProject(noConfig)
	sqlConnectionName := getInstance(project)
	sqlInstanceName := strings.Split(sqlConnectionName, ":")
	database := getDatabase(sqlInstanceName[2], project)
	doConnect(project, sqlConnectionName, sqlInstanceName[2], database, port, debug, direct, noIam, timeout, unixSocket)
}

// timeout > 0 のときは supervise サブコマンド経由で起動し、ttl 経過後に
// proxy を自動停止させる。0 なら従来どおり proxy を直接常駐させる。
func startProxyBackground(proxyArgs []string, timeout time.Duration) *exec.Cmd {
	var cmd *exec.Cmd
	if timeout > 0 {
		self, err := os.Executable()
		if err != nil {
			log.Fatal(err)
		}
		args := append([]string{"supervise", "--ttl", timeout.String(), "--"}, proxyArgs...)
		cmd = exec.Command(self, args...)
	} else {
		cmd = exec.Command("cloud-sql-proxy", proxyArgs...)
	}
	cmd.Stdout = os.Stdout
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	return cmd
}

func printAutoDisconnect(timeout time.Duration) {
	if timeout > 0 {
		color.Yellow("This connection will be automatically disconnected at %s (--timeout %s, set 0 to disable)", time.Now().Add(timeout).Format("15:04"), timeout)
	}
}

func doConnect(project, sqlConnectionName, instance, database string, port int, debug bool, direct bool, noIam bool, timeout time.Duration, unixSocket string) {
	if unixSocket != "" {
		doConnectUnixSocket(project, sqlConnectionName, instance, database, debug, noIam, unixSocket)
	} else {
		doConnectTCP(project, sqlConnectionName, instance, database, port, debug, direct, noIam, timeout)
	}
}

func doConnectTCP(project, sqlConnectionName, instance, database string, port int, debug bool, direct bool, noIam bool, timeout time.Duration) {
	if !direct {
		existingPort := findExistingProxy(sqlConnectionName)
		if existingPort > 0 {
			userName := getUser()
			dbTypeName := getdbTypeName(instance, project)
			boldGreen := color.New(color.FgGreen, color.Bold)
			color.Yellow("Already connected to %s on port %d", sqlConnectionName, existingPort)
			color.Blue("Can connect using:")
			if strings.Contains(dbTypeName, "POSTGRES") {
				_, _ = boldGreen.Printf("psql -h localhost -U %s -p %d -d %s\n", userName, existingPort, database)
			}
			if strings.Contains(dbTypeName, "MYSQL") {
				var re = regexp.MustCompile("@.*")
				_, _ = boldGreen.Printf("mysql --user=%s --password=`gcloud auth print-access-token` --enable-cleartext-plugin --host=127.0.0.1 --port=%d --database=%s\n", re.ReplaceAllString(userName, ""), existingPort, database)
			}
			return
		}
	}

	dbTypeName := getdbTypeName(instance, project)
	if port == 5432 && strings.Contains(dbTypeName, "MYSQL") {
		port = 3306
	}
	port = findAvailablePort(port)
	userName := getUser()

	iamFlag := " --auto-iam-authn"
	if noIam {
		iamFlag = ""
	}

	// color setting
	green := color.New(color.FgGreen)
	blue := color.New(color.FgBlue)
	boldGreen := green.Add(color.Bold)
	boldBlue := blue.Add(color.Bold)

	// connect instance
	if strings.Contains(dbTypeName, "POSTGRES") {
		if direct {
			var err error
			command := fmt.Sprintf("cloud-sql-proxy %s%s --private-ip --port=%d", sqlConnectionName, iamFlag, port)
			if debug {
				command = fmt.Sprintf("cloud-sql-proxy %s%s --debug --private-ip --port=%d", sqlConnectionName, iamFlag, port)
				color.Blue("[Debug Mode]\nThe following commands are executed in foreground.\n")
				_, _ = boldBlue.Printf("%s\n", command)
			}

			proxyCmd := exec.Command("bash", "-c", command)
			if debug {
				proxyCmd.Stderr = os.Stderr
				proxyCmd.Stdout = os.Stdout
			}
			err = proxyCmd.Start()
			if err != nil {
				log.Fatal(err)
			}

			log.Printf("Cloudsql proxy process is running in background, process_id: %d\n", proxyCmd.Process.Pid)

			// Retry logic to wait for the proxy to be ready.
			var conn net.Conn
			retryCount := 30
			for i := 0; i < retryCount; i++ {
				conn, err = net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), 500*time.Millisecond)
				if err == nil {
					_ = conn.Close()
					break
				}
				time.Sleep(500 * time.Millisecond)
			}

			if err != nil {
				log.Println("Failed to connect to cloud-sql-proxy. Killing proxy.")
				_ = proxyCmd.Process.Kill()
				log.Fatalf("Proxy connection error: %v", err)
			}

			psqlCommand := fmt.Sprintf("psql -h localhost -U %s -p %d -d %s", userName, port, database)

			psqlCmd := exec.Command("bash", "-c", psqlCommand)
			psqlCmd.Stdout = os.Stdout
			psqlCmd.Stderr = os.Stderr
			psqlCmd.Stdin = os.Stdin
			err = psqlCmd.Run()
			if err != nil {
				log.Println("psql command failed. Killing proxy.")
				_ = proxyCmd.Process.Kill()
				log.Fatal(err)
			}
			err = proxyCmd.Process.Kill()
			if err != nil {
				log.Fatal(err)
			}
		} else {
			command := fmt.Sprintf("cloud-sql-proxy %s%s --private-ip --quiet --port=%d", sqlConnectionName, iamFlag, port)
			if debug {
				command = fmt.Sprintf("cloud-sql-proxy %s%s --debug --private-ip --port=%d", sqlConnectionName, iamFlag, port)
				color.Blue("[Debug Mode]\nThe following commands are executed in foreground.\n")
				_, _ = boldBlue.Printf("%s\n", command)
				debug := exec.Command("bash", "-c", command)
				debug.Stdout = os.Stdout
				debug.Stderr = os.Stderr
				err := debug.Run()
				if err != nil {
					log.Fatal(err)
				}
			} else {
				proxyArgs := []string{sqlConnectionName}
				if !noIam {
					proxyArgs = append(proxyArgs, "--auto-iam-authn")
				}
				proxyArgs = append(proxyArgs, "--private-ip", "--quiet", "--port="+strconv.Itoa(port))
				cmd := startProxyBackground(proxyArgs, timeout)
				log.Printf("Cloudsql proxy process is running in background, process_id: %d\n", cmd.Process.Pid)
				printAutoDisconnect(timeout)
			}

			color.Blue("Can connect using:")
			_, _ = boldGreen.Printf("psql -h localhost -U %s -p %d -d %s\n", userName, port, database)
		}
	}
	if strings.Contains(dbTypeName, "MYSQL") {
		if debug {
			command := fmt.Sprintf("cloud-sql-proxy %s%s --private-ip --debug --port=%d", sqlConnectionName, iamFlag, port)
			color.Blue("[Debug Mode]\nThe following commands are executed in foreground.\n")
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

		port = findAvailablePort(port)
		mysqlProxyArgs := []string{sqlConnectionName}
		if !noIam {
			mysqlProxyArgs = append(mysqlProxyArgs, "--auto-iam-authn")
		}
		mysqlProxyArgs = append(mysqlProxyArgs, "--private-ip", "--quiet", "--port="+strconv.Itoa(port))
		cmd := startProxyBackground(mysqlProxyArgs, timeout)
		log.Printf("Cloudsql proxy process is running in background, process_id: %d\n", cmd.Process.Pid)
		printAutoDisconnect(timeout)

		color.Blue("Can connect using:")
		var re = regexp.MustCompile("@.*")
		_, _ = boldGreen.Printf("mysql --user=%s --password=`gcloud auth print-access-token` --enable-cleartext-plugin --host=127.0.0.1 --port=%d --database=%s\n", re.ReplaceAllString(userName, ""), port, database)
		// Temporarily commented out when database is selected because the connection is not possible due to permission issues.
		//_, _ = boldGreen.Printf("mysql --user=%s --password=`gcloud auth print-access-token` --enable-cleartext-plugin --host=127.0.0.1 --port=%d --database=%s\n", re.ReplaceAllString(userName, ""), port, database)
	}
}

func doConnectUnixSocket(project, sqlConnectionName, instance, database string, debug bool, noIam bool, unixSocket string) {
	dbTypeName := getdbTypeName(instance, project)
	userName := getUser()

	boldGreen := color.New(color.FgGreen, color.Bold)
	boldBlue := color.New(color.FgBlue, color.Bold)

	// 0700 で他ユーザーからの socket 接続を OS レベルで遮断する（--auto-iam-authn 前提のため必須）
	if err := os.MkdirAll(unixSocket, 0700); err != nil {
		log.Fatalf("Failed to create unix socket directory %s: %v", unixSocket, err)
	}
	if err := os.Chmod(unixSocket, 0700); err != nil {
		log.Fatalf("Failed to set permissions on unix socket directory %s: %v", unixSocket, err)
	}

	sanitizedName := strings.ReplaceAll(sqlConnectionName, ":", "_")
	socketPath := filepath.Join(unixSocket, sanitizedName)

	if _, err := os.Stat(socketPath); err == nil {
		if findExistingProxyProcess(sqlConnectionName) {
			color.Yellow("Already connected to %s via unix socket", sqlConnectionName)
			color.Blue("Can connect using:")
			if strings.Contains(dbTypeName, "POSTGRES") {
				_, _ = boldGreen.Printf("psql -h %s -U %s -d %s\n", filepath.Dir(socketPath), userName, database)
			}
			if strings.Contains(dbTypeName, "MYSQL") {
				var re = regexp.MustCompile("@.*")
				_, _ = boldGreen.Printf("mysql --user=%s --password=`gcloud auth print-access-token` --enable-cleartext-plugin --socket=%s --database=%s\n", re.ReplaceAllString(userName, ""), socketPath, database)
			}
			return
		}
		os.Remove(socketPath)
	}

	instanceArg := fmt.Sprintf("'%s?unix-socket-path=%s'", sqlConnectionName, socketPath)

	iamFlag := " --auto-iam-authn"
	if noIam {
		iamFlag = ""
	}

	if debug {
		command := fmt.Sprintf("cloud-sql-proxy %s%s --debug --private-ip", instanceArg, iamFlag)
		color.Blue("[Debug Mode]\nThe following commands are executed in foreground.\n")
		_, _ = boldBlue.Printf("%s\n", command)
		color.Blue("Can connect using:")
		if strings.Contains(dbTypeName, "POSTGRES") {
			_, _ = boldGreen.Printf("psql -h %s -U %s -d %s\n", unixSocket, userName, database)
		}
		if strings.Contains(dbTypeName, "MYSQL") {
			var re = regexp.MustCompile("@.*")
			_, _ = boldGreen.Printf("mysql --user=%s --password=`gcloud auth print-access-token` --enable-cleartext-plugin --socket=%s --database=%s\n", re.ReplaceAllString(userName, ""), socketPath, database)
		}
		debugCmd := exec.Command("bash", "-c", command)
		debugCmd.Stdout = os.Stdout
		debugCmd.Stderr = os.Stderr
		err := debugCmd.Run()
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	command := fmt.Sprintf("cloud-sql-proxy %s%s --private-ip --quiet", instanceArg, iamFlag)
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = os.Stdout
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Cloudsql proxy process is running in background (unix socket), process_id: %d\n", cmd.Process.Pid)

	color.Blue("Can connect using:")
	if strings.Contains(dbTypeName, "POSTGRES") {
		_, _ = boldGreen.Printf("psql -h %s -U %s -d %s\n", filepath.Dir(socketPath), userName, database)
	}
	if strings.Contains(dbTypeName, "MYSQL") {
		var re = regexp.MustCompile("@.*")
		_, _ = boldGreen.Printf("mysql --user=%s --password=`gcloud auth print-access-token` --enable-cleartext-plugin --socket=%s --database=%s\n", re.ReplaceAllString(userName, ""), socketPath, database)
	}
}
