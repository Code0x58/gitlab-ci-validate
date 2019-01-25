workflow "Test and Publish" {
  on = "push"
  resolves = "Release"
}

action "Test" {
  uses = "./.github/actions/go-check"
  args = "go-check"
}

action "Release" {
  needs = ["Test"]
  uses = "./.github/actions/go-release"
  args = "go-release"
  secrets = ["GITHUB_TOKEN"]
}
