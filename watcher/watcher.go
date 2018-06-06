package watcher

import (
	"fmt"
	"github.com/0xAX/notificator"
	"github.com/rjeczalik/notify"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var notifier *notificator.Notificator

type WatchGroup struct {
	BaseDir        string `mapstructure:"base_dir"`
	TestDir        string `mapstructure:"test_dir"`
	CodeDir        string `mapstructure:"code_dir"`
	TestRunner     string `mapstructure:"test_runner"`
	Name           string `json:"name"`
	TestRegex      string `mapstructure:"test_regex"`
	WatchExtension string `mapstructure:"watch_extension"`
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

func FindTest(testDir string, path string, testRegexStr string) (fileList []string, err error) {
	filename := filepath.Base(path)
	// if the modified file is in the testDir path assume it's a test itself
	// and return it.
	if strings.HasPrefix(filepath.Dir(path), testDir) {
		fileList = []string{path}
		return
	}

	// replace the <FILE> placeholder in regex string with the actual file name
	testRegex_ := strings.Replace(testRegexStr, "<FILE>", filenameNoExt(filename), 1)
	testRegex := regexp.MustCompile(testRegex_)

	err = filepath.Walk(testDir, func(path string, f os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
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
		failMessage := fmt.Sprintf("%s:\n%s", w.Name, filepath.Base(testPath))
		notifier.Push("GOTESTIT FAIL", failMessage, "", notificator.UR_CRITICAL)
	} else {
		succMessage := fmt.Sprintf("%s:\n%s", w.Name, filepath.Base(testPath))
		notifier.Push("GOTESTIT Pass", succMessage, "", notificator.UR_NORMAL)
	}
}

func (w WatchGroup) waitForTests(c chan ChangedFile) {
	fmt.Printf("Watching for changes in %s project\n", w.Name)
	watchDir := w.CodeDir + "..."

	evChan := make(chan notify.EventInfo)
	if err := notify.Watch(watchDir, evChan, notify.InCloseWrite, notify.InMovedTo); err != nil {
		log.Fatalln(err)
	}

	for ev := range evChan {
		c <- ChangedFile{wg: &w, path: ev.Path()}
	}
	defer notify.Stop(evChan)
}

func getStringConfig(key string, v map[interface{}]interface{}) string {
	value, ok := v[key].(string)
	if ok == false {
		log.Fatalln(fmt.Errorf("Missing required config value: %s \n", key))
	}
	return value
}

func Watch(ws []WatchGroup) {
	notifier = notificator.New(notificator.Options{
		DefaultIcon: "icon/default.png",
		AppName:     "GOTESTIT",
	})

	chFiles := make(chan ChangedFile, 10)

	for _, w := range ws {
		go w.waitForTests(chFiles)
	}

	for c := range chFiles {
		if filepath.Ext(c.path) == c.wg.WatchExtension {
			var fileList []string

			fileList, _ = FindTest(c.wg.TestDir, c.path, c.wg.TestRegex)

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
