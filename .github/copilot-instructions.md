# Words for the Agents
This will be updated regularly to improve code quality.
If you see a file that does not meet the current standards, please take a moment to update the file.


# Code Quality
## Config Files
- All config files should have comments if they allow for comments.
## Go
- All Go code should have comments for documentation.
- Always use `go run` to test your changes, don't compile.

# Don't Doxx Users
If you're writing a file, paths and data is very rarely hard coded.
Do not hard code paths or data unless absolutely necessary.
The exception to this is when doing something like GitOps, where you mostly need to ensure secrets aren't in the repo.
