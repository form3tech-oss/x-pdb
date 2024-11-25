package tests

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func getProjectDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return wd, err
	}
	wd = strings.Replace(wd, "/tests", "", -1)
	return wd, nil
}

func runCmd(cmd *exec.Cmd) ([]byte, error) {
	dir, _ := getProjectDir()
	cmd.Dir = dir

	if err := os.Chdir(cmd.Dir); err != nil {
		fmt.Printf("chdir dir: %s\n\n", err)
	}

	cmd.Env = append(os.Environ(), "GO111MODULE=on")
	command := strings.Join(cmd.Args, " ")
	//fmt.Fprintf(GinkgoWriter, "running: %s\n", command)
	fmt.Printf("running: %s\n\n", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, fmt.Errorf("%s failed with error: (%v) %s", command, err, string(output))
	}

	return output, nil
}
