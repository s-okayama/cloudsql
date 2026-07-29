package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/manifoldco/promptui"
)

func selectInstanceEntry() string {
	str := listInstance()

	prompt := promptui.Select{
		Label:  "Select Instance to disconnect",
		Items:  str,
		Stdout: NoBellStdout,
	}

	_, result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(0)
		return ""
	}

	fmt.Printf("You choose %q\n", result)
	return result
}

func unixSocketPathFromEntry(entry string) string {
	m := regexp.MustCompile(`unix-socket-path=([^' ]+)`).FindStringSubmatch(entry)
	if m == nil {
		return ""
	}
	return m[1]
}

func disconnectInstance(all bool) {
	if all {
		command := "ps aux | grep 'cloud-sql-proxy' | grep -v grep | awk '{print $2}' | xargs kill -9"
		cmd := exec.Command("bash", "-c", command)
		err := cmd.Run()
		if err != nil {
			// This can happen if no processes are found, which is not a fatal error.
			log.Println("No cloud-sql-proxy processes found to disconnect, or an error occurred.")
			return
		}
		log.Println("All cloud-sql-proxy instances disconnected.")
		return
	}

	entry := selectInstanceEntry()
	if socketPath := unixSocketPathFromEntry(entry); socketPath != "" {
		// unix socket 接続はポートを持たないため、socket パスでプロセスを特定して kill する
		command := fmt.Sprintf("ps aux | grep cloud-sql-proxy | grep -v grep | grep '%s' | awk '{print $2}' | xargs kill -9", socketPath)
		if err := exec.Command("bash", "-c", command).Run(); err != nil {
			log.Fatal(err)
		}
		_ = os.Remove(socketPath)
		log.Printf("Instance disconnected")
		return
	}

	res1 := strings.Split(entry, "=")
	port := res1[len(res1)-1]
	command := fmt.Sprintf("lsof -i tcp:%s | grep LISTEN | awk '{print $2}' | xargs kill -9", port)
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Instance disconnected")
}
