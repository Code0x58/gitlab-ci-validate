## gitlab-ci-validate
This tool uses GitLab's CI [config validation API endpoint](https://docs.gitlab.com/ce/api/lint.html) to validate local config files.

If you don't want to use the command line, you can paste your config into `https://gitlab.com/<your project>/-/ci/lint` [[ref](https://docs.gitlab.com/ee/ci/yaml/#validate-the-gitlab-ciyml)]

### Usage
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
 - 2 if there was any _SOFT FAIL_ and no _HARD FAIL_

### Using private GitLab host
You can also use a private GitLab host both as a flag or as an environment variable.
The following are equivalent.

```
gitlab-ci-validate --host=http://user:pass@127.0.0.1:8080 .gitlab-ci.yml
```

```
export GITLAB_HOST=http://user:pass@127.0.0.1:8080
gitlab-ci-validate .gitlab-ci.yml
```

The flag has precedence over the environment variable.
When not specified the host used is by default `https://gitlab.com`

### Installation
You can either use a premade binary from the [releases page](https://github.com/Code0x58/gitlab-ci-validate/releases) or you can install it using `go get`:
```sh
go get -u github.com/Code0x58/gitlab-ci-validate
```

#### Usage with Docker containers
You can use the Dockerfile to build your own image or use the pre-built version available at the [Gitlab container registry](https://gitlab.com/comedian780/docker-gitlab-ci-validate/container_registry).

You can run tests against the gitlab.com endpoint:  
If no parameter is given the container will look for a file called `.gitlab-ci.yml`
```sh
docker run -i --rm \
-v ${PWD}/.gitlab-ci.yml:/yaml/.gitlab-ci.yml \
registry.gitlab.com/comedian780/docker-gitlab-ci-validate
```

You can run tests against a self hosted Gitlab instance with custom filenames:  
Set the credentials and URL via the `GITLAB_HOST` environment variable  
```sh
docker run -i --rm \
-e GITLAB_HOST=https://GITLAB_USER:GITLAB_PW@your.gitlab.server
-v ${PWD}:/yaml \
-v /additional/folder/.additional.yml:/yaml/.additional.yml \
registry.gitlab.com/comedian780/docker-gitlab-ci-validate custom.yml .files.yaml .additional.yml
```

You can also test all YAML files inside a directory (this also includes YAML files in subdirectories):
```sh
find . -type f -regex ".*\.\(yaml\|yml\|YAML\|YML\)" | xargs -I {}
docker run -i --rm \
-v ${PWD}:/yaml \
registry.gitlab.com/comedian780/docker-gitlab-ci-validate {}
```
