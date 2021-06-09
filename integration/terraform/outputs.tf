# Generated by: tyk-ci/wf-gen
# Generated on: Wed  9 Jun 01:05:50 UTC 2021

# Generation commands:
# ./pr.zsh -title amd64 only for rhel -branch releng/manifests -repos tyk-identity-broker -base releng/manifests -p
# m4 -E -DxREPO=tyk-identity-broker


data "terraform_remote_state" "integration" {
  backend = "remote"

  config = {
    organization = "Tyk"
    workspaces = {
      name = "base-prod"
    }
  }
}

output "tyk-identity-broker" {
  value = data.terraform_remote_state.integration.outputs.tyk-identity-broker
  description = "ECR creds for tyk-identity-broker repo"
}

output "region" {
  value = data.terraform_remote_state.integration.outputs.region
  description = "Region in which the env is running"
}
