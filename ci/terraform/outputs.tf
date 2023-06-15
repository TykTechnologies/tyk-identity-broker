
# Generated by: gromit policy
# Generated on: Thu Jun 15 13:17:16 UTC 2023



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
