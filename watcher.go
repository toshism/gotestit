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
var config_file string = "config.json"

// the file extension to monitor for changes
// all other files will be ignored
var file_ext string = ".py"

type WatchGroup struct {
	BaseDir    string `json:"base_dir"`
	CodeDir    string `json:"code_dir"`
	TestDir    string `json:"test_dir"`
	TestRunner string `json:"test_runner"`
	Name       string `json:"name"`
}

type ChangedFile struct {
	w WatchGroup
	f string
}

func loadConfig(filename string) ([]WatchGroup, error) {
	watch_groups := make([]WatchGroup, 0)

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return watch_groups, err
	}
	err = json.Unmarshal(bytes, &watch_groups)
	if err != nil {
		return watch_groups, err
	}
	return watch_groups, nil
}

func findTest(testDir string, filename string) ([]string, error) {
	fileList := []string{}
	err := filepath.Walk(testDir, func(path string, f os.FileInfo, err error) error {
		if filename == filepath.Base(path) {
			fileList = append(fileList, path)
		}
		return nil
	})

	if err != nil {
		return fileList, err
	}

	return fileList, nil
}

func run_test(test_path string, w WatchGroup) {
	fmt.Printf("\nFile saved in %s project\n", w.Name)
	fmt.Printf("Running test: %s\n", test_path)
	cmd := exec.Command(w.TestRunner, test_path)
	cmd.Dir = w.BaseDir

	var out bytes.Buffer
	// nose prints everything to stderr?
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	// it would be cooler to stream the test result out put. see below
	// https://nathanleclaire.com/blog/2014/12/29/shelled-out-commands-in-golang/
	fmt.Printf(out.String())
}

func wait_for_tests(w WatchGroup, c chan ChangedFile) {
	fmt.Printf("Watching for changes in %s project\n", w.Name)
	watch_dir := w.CodeDir + "..."

	t := make(chan notify.EventInfo)
	if err := notify.Watch(watch_dir, t, notify.Write); err != nil {
		log.Fatal(err)
	}

	for i := range t {
		c <- ChangedFile{w: w, f: i.Path()}
	}
	defer notify.Stop(t)

}

func main() {
	var ws []WatchGroup
	ws, _ = loadConfig(config_file)

	c := make(chan ChangedFile)

	for _, w := range ws {
		go wait_for_tests(w, c)
	}

	for i := range c {
		path := i.f
		w := i.w
		if filepath.Ext(path) == file_ext {
			fileList, _ := findTest(w.TestDir, "test_"+filepath.Base(path))
			if len(fileList) > 0 {
				for _, f := range fileList {
					run_test(f, w)
				}
			} else {
				fmt.Printf("Couldn't find any tests for %s. You should write some.\n", path)
			}
		}
	}
}
