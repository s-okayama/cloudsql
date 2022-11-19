package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/manifoldco/promptui"
)

func getPort() string {
	str := listInstance()

	prompt := promptui.Select{
		Label:  "Select Instance to disconnect",
		Items:  str,
		Stdout: NoBellStdout,
	}

	_, result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
		return ""
	}

	fmt.Printf("You choose %q\n", result)
	res1 := strings.Split(result, ":")
	port := res1[len(res1)-1]
	return port
}

func disconnectInstance() {
	port := getPort()
	command := fmt.Sprintf("lsof -i tcp:%s | grep LISTEN | awk '{print $2}' | xargs kill -9", port)
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = os.Stdout
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Instance disconnected")
}
