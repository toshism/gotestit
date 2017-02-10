package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/rjeczalik/notify"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// location of the config file
var configFile = "config.json"

// the file extension to monitor for changes
// all other files will be ignored
var fileExt = ".py"

type WatchGroup struct {
	BaseDir    string `json:"base_dir"`
	CodeDir    string `json:"code_dir"`
	TestDir    string `json:"test_dir"`
	TestRunner string `json:"test_runner"`
	Name       string `json:"name"`
}

type ChangedFile struct {
	wg   *WatchGroup
	path string
}

func loadConfig(filename string) (watchGroups []WatchGroup, err error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	err = json.Unmarshal(bytes, &watchGroups)
	return
}

func findTest(testDir string, filename string) (fileList []string, err error) {
	err = filepath.Walk(testDir, func(path string, f os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if filename == filepath.Base(path) {
			fileList = append(fileList, path)
		}
		return nil
	})
	return
}

func (w WatchGroup) runTest(testPath string) {
	fmt.Printf("\nFile saved in %s project\n", w.Name)
	fmt.Printf("Running test: %s\n", testPath)
	cmd := exec.Command(w.TestRunner, testPath)
	cmd.Dir = w.BaseDir

	var out bytes.Buffer
	// nose prints everything to stderr?
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
	// it would be cooler to stream the test result out put. see below
	// https://nathanleclaire.com/blog/2014/12/29/shelled-out-commands-in-golang/
	fmt.Printf(out.String())
}

func (w WatchGroup) waitForTests(c chan ChangedFile) {
	fmt.Printf("Watching for changes in %s project\n", w.Name)
	watchDir := w.CodeDir + "..."

	evChan := make(chan notify.EventInfo)
	if err := notify.Watch(watchDir, evChan, notify.Write); err != nil {
		log.Fatal(err)
	}

	for ev := range evChan {
		c <- ChangedFile{wg: &w, path: ev.Path()}
	}
	defer notify.Stop(evChan)
}

func main() {
	var ws []WatchGroup
	ws, _ = loadConfig(configFile)

	chFiles := make(chan ChangedFile)

	for _, w := range ws {
		go w.waitForTests(chFiles)
	}

	for c := range chFiles {
		if filepath.Ext(c.path) == fileExt {
			fileList, _ := findTest(c.wg.TestDir, "test_"+filepath.Base(c.path))
			if len(fileList) > 0 {
				for _, f := range fileList {
					c.wg.runTest(f)
				}
			} else {
				fmt.Printf("Couldn't find any tests for %s. You should write some.\n", c.path)
			}
		}
	}
}
