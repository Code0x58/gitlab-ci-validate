# gitlab-ci-validate

This tool uses GitLab's CI [config validation API endpoint](https://docs.gitlab.com/ce/api/lint.html) to validate local config files.

If you don't want to use the command line, you can paste your config into `https://gitlab.com/<your project>/-/ci/lint` [[ref](https://docs.gitlab.com/ee/ci/yaml/#validate-the-gitlab-ciyml)]

## Usage

> :warning: Since GitLab 13.7.7 (2021-02-11) authentication is required, so you will need to use `-token=$ACCESS_TOKEN` or `--host=http://$USERNAME:$PASSWORD@gitlab.com`

> :warning: Since GitLab 16.0 (2023-05-17) project ID is required, so you will need to use `-project-id=$PROJECT_ID`

One or more `.gitlab-ci.yml` are passed as arguments on the command line. Any errors will result in a non-zero exit code. The filename must end in `.yml` to pass, but doesn't have to be `.gitlab-ci.yml`.

An access token must be provided in order to authenticate with the gitlab API. You can see your access tokens through [your profile settings](https://gitlab.com/-/profile/personal_access_tokens). The token must have at least the "api" and "read_api" scopes.

```text
$ gitlab-ci-validate --token=ACCESS_TOKEN ./good.yml ./maybe-good.yml ./bad.yml
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

## Using private GitLab host

You can also use a private GitLab host both as a flag or as an environment variable.
The following are equivalent.

```shell
gitlab-ci-validate -token=$ACCESS_TOKEN -host=http://$USERNAME:$PASSWORD@127.0.0.1:8080 -project-id=1234 .gitlab-ci.yml
```

```shell
export GITLAB_HOST=http://user:pass@127.0.0.1:8080
export GITLAB_TOKEN=$ACCESS_TOKEN
export GITLAB_PROJECT_ID=1234
gitlab-ci-validate .gitlab-ci.yml
```

The flag has precedence over the environment variable.
When not specified the host used is by default `https://gitlab.com`

## Installation

You can either use a premade binary from the [releases page](https://github.com/Code0x58/gitlab-ci-validate/releases) or you can install it using `go get`:

```sh
go get -u github.com/Code0x58/gitlab-ci-validate
```

### Usage with Docker containers

You can use the Dockerfile to build your own image, or use the pre-built version available at the [GitHub Container Registry](https://github.com/Code0x58/gitlab-ci-validate/packages/) - you will need to be logged in first (see [docs](https://docs.github.com/en/packages/guides/configuring-docker-for-use-with-github-packages#authenticating-to-github-packages)).

The default argument given to `gitlab-ci-validate` in the container is `.gitlab-ci.yml`, so the following will check that file from the current working directory:

```shell
docker run -i --rm \
    -v "${PWD}":/yaml \
    ghcr.io/code0x58/gitlab-ci-validate/gitlab-ci-validate:$VERSION
```
