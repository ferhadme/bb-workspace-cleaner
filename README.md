# Bitbucket Workspace Cleaner

### Description
Command line tool for removing all branches of repositories of organization that last updated 3 months ago

### Compilation
```sh
$ go build
```

### Usage
```sh
$ ./bb-workspace-cleaner <user> <organization> <months>
```
*user* is username of BitBucket user

*organization* is name of the organization owns repositories

*months* is amount of month repository last updated

### Pass.txt file
Program uses [App passwords](https://support.atlassian.com/bitbucket-cloud/docs/app-passwords/) for authenticating BitBucket endpoints.
Generated App password should be written to **pass.txt** file
