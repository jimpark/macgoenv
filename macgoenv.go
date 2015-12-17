package main

import (
	"bufio"
	"fmt"
	"github.com/DHowett/go-plist"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
)

func isDir(path string) bool {
	stat, err := os.Stat(path)
	if err == nil {
		return stat.IsDir()
	}
	return false
}

func isFile(path string) bool {
	stat, err := os.Stat(path)
	if err == nil {
		return !stat.IsDir()
	}
	return false
}

type envPlist struct {
	Label            string   `plist:"Label"`
	ProgramArguments []string `plist:"ProgramArguments"`
	RunAtLoad        bool     `plist:"RunAtLoad"`
}

func readEnvPlistContent(path string) (result map[string]string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}

	defer f.Close()

	dec := plist.NewDecoder(f)
	data := new(envPlist)
	err = dec.Decode(data)

	if err != nil {
		return
	}

	r, err := regexp.Compile(`launchctl\s+setenv\s+(\w+)\s+(.*)`)
	if err != nil {
		log.Fatal(err)
	}

	result = make(map[string]string)
	for _, s := range data.ProgramArguments {
		m := r.FindStringSubmatch(s)
		if len(m) == 3 {
			result[m[1]] = m[2]
		}
	}

	return
}

func createEnvPlistContent(vars map[string]string) *envPlist {
	result := new(envPlist)
	result.Label = "my.startup"
	result.RunAtLoad = true
	result.ProgramArguments = []string{"sh", "-c"}

	for key, value := range vars {
		result.ProgramArguments = append(result.ProgramArguments, fmt.Sprintf("launchctl setenv %s %s", key, value))
	}
	return result
}

func main() {
	if runtime.GOOS != "darwin" {
		fmt.Println("Error: this program only runs on the Mac.")
		return
	}

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	macosxdir := filepath.Join(usr.HomeDir, "Library", "LaunchAgents")

	if !isDir(macosxdir) {
		err := os.MkdirAll(macosxdir, 0700)
		if err != nil {
			fmt.Printf("Error: Cannot create directory %s.\n", macosxdir)
			log.Fatal(err)
		}
	}

	var env map[string]string

	plistPath := filepath.Join(macosxdir, "environment.plist")

	if isFile(plistPath) {
		env, err = readEnvPlistContent(plistPath)
		if err != nil {
			log.Fatal(err)
		}
	}

	if env == nil {
		env = map[string]string{
			"GOPATH": os.Getenv("GOPATH"),
		}
	}

	fmt.Printf("GOPATH = %s\n", env["GOPATH"])

	if len(env["GOPATH"]) == 0 {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter desired Go path (GOPATH): ")
		env["GOPATH"], _ = reader.ReadString('\n')
	}

	contents := createEnvPlistContent(env)

	f, err := os.Create(plistPath)
	if err != nil {
		fmt.Printf("Error: could not create '%s' file.\n", plistPath)
		log.Fatal(err)
	}
	defer f.Close()
	encoder := plist.NewEncoderForFormat(f, plist.AutomaticFormat)

	if err != nil {
		log.Fatal(err)
	}
	encoder.Encode(contents)

	fmt.Printf("Created '%s' file.\n", plistPath)

	fmt.Println("Running launchctl to apply settings.")
	cmd := exec.Command("launchctl", "unload", plistPath)
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	cmd = exec.Command("launchctl", "load", plistPath)
	if err != nil {
		log.Fatal(err)
	}
}
