workflow "Check and Publish" {
  on = "push"
  resolves = "Release"
}

action "Check" {
  uses = "./.github/actions/go-check"
  args = "go-check"
}

action "Release" {
  needs = ["Check"]
  uses = "./.github/actions/go-release"
  args = "go-release"
  secrets = ["GITHUB_TOKEN"]
}
