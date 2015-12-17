package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
)

func isDir(path string) bool {
	stat, err := os.Stat(path)
	if err == nil {
		return stat.IsDir()
	}
	return false
}

func createEnvPlistContent(vars map[string]string) string {
	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
`)

	for key, value := range vars {
		buf.WriteString(fmt.Sprintf("\t<key>%s</key>\n", key))
		buf.WriteString(fmt.Sprintf("\t<string>%s</string>\n", value))
	}

	buf.WriteString("</dict>\n</plist>\n")
	return buf.String()
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

	macosxdir := filepath.Join(usr.HomeDir, ".MacOSX")

	if !isDir(macosxdir) {
		err := os.Mkdir(macosxdir, 0700)
		if err != nil {
			fmt.Printf("Error: Cannot create directory %s.\n", macosxdir)
			log.Fatal(err)
		}
	}

	env := map[string]string{
		"GOPATH": os.Getenv("GOPATH"),
	}

	if len(env["GOPATH"]) == 0 {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter desired Go path (GOPATH): ")
		env["GOPATH"], _ = reader.ReadString('\n')
	}

	contents := createEnvPlistContent(env)

	plistPath := filepath.Join(macosxdir, "environment.plist")
	f, err := os.Create(plistPath)
	if err != nil {
		fmt.Printf("Error: could not create '%s' file.\n", plistPath)
		log.Fatal(err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	_, err = w.WriteString(contents)

	if err != nil {
		log.Fatal(err)
	}
	w.Flush()
	fmt.Printf("Created '%s' file.\n", plistPath)
	fmt.Println("Log out and log back in to take effect.")
}
