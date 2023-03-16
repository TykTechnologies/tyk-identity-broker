terraform {

  #Being used until TFCloud can be used
  backend "remote" {
    hostname     = "app.terraform.io"
    organization = "Tyk"
    workspaces {
      name = "repo-policy-tyk-identity-broker"
    }
  }

  required_providers {
    github = {
      source  = "integrations/github"
      version = "5.16.0"
    }
  }
}

provider "github" {
  owner = "TykTechnologies"
}

module "tyk-identity-broker" {
  source               = "./modules/github-repos"
  repo                 = "tyk-identity-broker"
  description          = "Tyk Authentication Proxy for third-party login"
  default_branch       = "master"
  topics                      = []
  visibility                  = "public"
  wiki                        = true
  vulnerability_alerts        = true
  squash_merge_commit_message = "COMMIT_MESSAGES"
  squash_merge_commit_title   = "COMMIT_OR_PR_TITLE"
  release_branches     = [
{ branch    = "master",
	reviewers = "1",
	convos    = "false",
	required_tests = ["1.16"]},
{ branch    = "release-1.2",
	reviewers = "0",
	convos    = "false",
	required_tests = ["1.15"]},
{ branch    = "release-1.3",
	reviewers = "0",
	convos    = "false",
	required_tests = ["1.16"]},
{ branch    = "release-1.4",
    reviewers = "0",
    convos    = "false",
    required_tests = ["1.16"]},
]
}