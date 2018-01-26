# gotestit
Watch for file changes and run corresponding tests

This is a toy project. I was tired of using guard and jacking 
with ruby version requirements to test a python project.

Also I just wanted to play around with Go.

**Update:** I've been using this for a fair amount of time now and it's actually
pretty solid.

Supports watching multiple projects simultaneously.

## Current status
Works for me... 

- ~I've only used it with nose test runner.~ Works with nose and pytest
- It expects tests to be name `test_` followed by the filename the test covers.
- It does not expect any specific directory structure.

## Setup
config file: `~/.gotestit.json`
see `gotestit.json` for an example
