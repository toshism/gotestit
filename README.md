# gotestit
Watch for file changes and run corresponding tests. That's it.

I was tired of using guard and jacking with ruby version requirements to test a
python project.

Also I just wanted to play around with Go.

The primary goal is to keep this simple. When a file is saved, find the
corresponding test file and run only those tests.

## Features
* Supports watching multiple projects simultaneously
* Regex for finding test files
* Does not eat all of your resources (looking at you xdist)
* Notification of pass/fail
* Only runs tests for the file you are editing*
* Theoretically supports any language

*see Limitations

## Limitations
* Tests must be in a separate directory
* Only supports a single test directory per project
* Test files must be named so that you can define a regex to match based on edited file name

## Configuration
Config is pretty simple. Define a few relevant paths (base directory, code,
tests), file type to watch for changes, set the test runner command, and give it
a name. You can optionally define a regex to match the changed file name to the
test file name.

Config file can be JSON, TOML, YAML, HCL, or Java properties and is named `gotestit` 
with the appropriate extension for whatever format you are using.

Search paths for the config file are (in order of precedence):
* $HOME/.config
* $HOME
* .

### Example Config file
YAML
```yaml
watch_extension: ".py"
test_regex: "test_<FILE>"
projects:
  - name: "Everything Scatter"
    base_dir: "/home/toshism/projects/es/"
    test_dir: "/home/toshism/projects/es/tests/"
    code_dir: "/home/toshism/projects/es/src/"
    test_runner: "/home/toshism/.virtualenvs/es/bin/pytest"
  - name: "Cosmic Slop"
    base_dir: "/home/toshism/projects/cosmic_slop/"
    test_dir: "/home/toshism/projects/cosmic_slop/tests/"
    code_dir: "/home/toshism/projects/cosmic_slop/src/"
    test_runner: "/home/toshism/.virtualenvs/cosmic_slop/bin/nose"
```
