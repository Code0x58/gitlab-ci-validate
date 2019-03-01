## gitlab-ci-validate
This tool uses GitLab's CI [config validation API endpoint](https://docs.gitlab.com/ce/api/lint.html) to validate local config files.

If you don't want to use the command line, you can paste your config into `https://gitlab.com/<your project>/-/ci/lint` [[ref](https://docs.gitlab.com/ee/ci/yaml/#validate-the-gitlab-ciyml)]

### Use
One or more `.gitlab-ci.yml` are passed as arguments on the command line. Any errors will result in a non-zero exit code. The filename must end in `.yml` to pass, but doesn't have to be `.gitlab-ci.yml`.
```text
$ gitlab-ci-validate ./good.yml ./maybe-good.yml ./bad.yml
PASS: ./good.yml
SOFT FAIL: ./maybe-good.yml
 - Post https://gitlab.com/api/v4/ci/lint: dial tcp: lookup gitlab.com on 127.0.0.53:53: read udp 127.0.0.1:41487->127.0.0.53:53: i/o timeout
HARD FAIL: ./bad.yml
 - jobs:storage config contains unknown keys: files
```

Each input file will be validated and one of 3 results will be printed for it:

 - _PASS_ - the file passed all checks
 - _SOFT FAIL_ - the file is acessable and contains valid YAML, but there was an error contacting the validation API
 - _HARD FAIL_ - the file failed any checks

The exit code will be:

 - 0 if all files are valid (all _PASS_)
 - 1 if any files are invalid (any _HARD FAIL_)
 - 2 if there were no _HARD FAIL_​s but any _SOFT FAIL_​s


### Installation
You can either use a premade binary from the [releases page](https://github.com/Code0x58/gitlab-ci-validate/releases) or you can install it using `go get`:
```sh
go get -u github.com/Code0x58/gitlab-ci-validate
```
