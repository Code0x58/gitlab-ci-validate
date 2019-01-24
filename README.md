## gitlab-ci-validate
Use GitLab's CI config validation API endpoint against local files.

### Use
One or more `.gitlab-ci.yml` are passed as arguments on the command line. Any errors will result in a non-zero exit code.
```
gitlab-ci-validate files...
```

### Installation
Until there are binary releases, you can get the tool using:
```sh
go get -u github.com/Code0x58/gitlab-ci-validate
```
