# gotestit
Watch for file changes and run corresponding tests

This is a toy project. I was tired of using guard and jacking 
with ruby version requirements to test a python project.

Also I just wanted to play around with Go.

Supports watching multiple projects simultaneously.

## Current status
Works for me... 

- I've only used it with nose test runner.
- It expects tests to be name `test_` followed by the filename the test covers.
- It does not expect any specific directory structure.
