terraform {
  required_providers {
    vault = {
      source  = "hashicorp/vault"
      version = "3.24.0"
    }
  }
}

provider "vault" {
  address = "http://localhost:8200"
}

# Put your Terraform code here
resource "vault_mount" "private_keys" {
  path        = "private_keys"
  type        = "transit"
  description = "transit secrect backend to store private keys"
}

resource "vault_policy" "remote_signer" {
  name   = "remote_signer"
  policy = file("policies/remote_signer.hcl")
}

resource "vault_auth_backend" "private_keys" {
  type = "approle"
  path = "private_keys"
}

resource "vault_approle_auth_backend_role" "remote_signer" {
  backend        = vault_auth_backend.private_keys.path
  role_name      = "remote_signer"
}

data "vault_approle_auth_backend_role_id" "remote_signer" {
  backend   = vault_auth_backend.private_keys.path
  role_name = vault_approle_auth_backend_role.remote_signer.role_name
}

output "role_id" {
  value = data.vault_approle_auth_backend_role_id.remote_signer.role_id
}
