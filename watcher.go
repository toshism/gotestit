package main

import (
	"fmt"
	"github.com/rjeczalik/notify"
	"github.com/spf13/viper"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var ws []WatchGroup
var fileExt string
var testRegexStr string

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

func filenameNoExt(s string) string {
	s = filepath.Base(s)
	n := strings.LastIndexByte(s, '.')
	if n >= 0 {
		return s[:n]
	}
	return s
}

func findTest(testDir string, filename string) (fileList []string, err error) {
	err = filepath.Walk(testDir, func(path string, f os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		testRegex_ := strings.Replace(testRegexStr, "<FILE>", filenameNoExt(filename), 1)
		testRegex := regexp.MustCompile(testRegex_)
		if testRegex.MatchString(filenameNoExt(path)) {
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
	cmd.Stderr = os.Stdout
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		failMessage := fmt.Sprintf("GOTESTIT FAIL:\n%s", filepath.Base(testPath))
		sendNotify := exec.Command("/usr/bin/notify-send", "-u", "critical", "-t", "3000", failMessage)
		sendNotify.Run()
	} else {
		succMessage := fmt.Sprintf("GOTESTIT Pass:\n%s", filepath.Base(testPath))
		sendNotify := exec.Command("/usr/bin/notify-send", "-h", "string:bgcolor:#00cc00", "-t", "3000", succMessage)
		sendNotify.Run()
	}
}

func (w WatchGroup) waitForTests(c chan ChangedFile) {
	fmt.Printf("Watching for changes in %s project\n", w.Name)
	watchDir := w.CodeDir + "..."

	evChan := make(chan notify.EventInfo)
	if err := notify.Watch(watchDir, evChan, notify.InCloseWrite, notify.InMovedTo); err != nil {
		log.Fatal(err)
	}

	for ev := range evChan {
		c <- ChangedFile{wg: &w, path: ev.Path()}
	}
	defer notify.Stop(evChan)
}

func getStringConfig(key string, v map[interface{}]interface{}) string {
	value, ok := v[key].(string)
	if ok == false {
		panic(fmt.Errorf("Missing required config value: %s \n", key))
	}
	return value
}

func init() {
	viper.SetConfigName("gotestit")
	viper.AddConfigPath("$HOME/.config")
	viper.AddConfigPath("$HOME")
	viper.AddConfigPath(".")

	// default config values
	viper.SetDefault("test_regex", "<FILE>")

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	fileExt = viper.GetString("watch_extension")
	testRegexStr = viper.GetString("test_regex")

	te := viper.Get("projects")
	test, _ := te.([]interface{})
	for _, value := range test {
		v, _ := value.(map[interface{}]interface{})
		wg := WatchGroup{
			BaseDir:    getStringConfig("base_dir", v),
			CodeDir:    getStringConfig("code_dir", v),
			TestDir:    getStringConfig("test_dir", v),
			TestRunner: getStringConfig("test_runner", v),
			Name:       getStringConfig("name", v),
		}
		ws = append(ws, wg)
	}

}

func main() {
	chFiles := make(chan ChangedFile, 10)

	for _, w := range ws {
		go w.waitForTests(chFiles)
	}

	for c := range chFiles {
		if filepath.Ext(c.path) == fileExt {
			var fileList []string

			if strings.HasPrefix(filepath.Dir(c.path), c.wg.TestDir) {
				// if the saved file is in the test directory itself,
				// assume it's a test and attempt to run it
				fileList = []string{c.path}
			} else {
				// search for a test that matches the test_regex
				fileList, _ = findTest(c.wg.TestDir, filepath.Base(c.path))
			}

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
