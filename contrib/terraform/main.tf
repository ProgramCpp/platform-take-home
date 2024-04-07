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

resource "vault_auth_backend" "remote_signer" {
  type = "approle"
  path = "private_keys"
}

resource "vault_approle_auth_backend_role" "remote_signer" {
  backend        = vault_auth_backend.remote_signer.path
  role_name      = "remote_signer"
  token_policies = ["remote_signer"]
}

data "vault_approle_auth_backend_role_id" "remote_signer" {
  backend   = vault_auth_backend.remote_signer.path
  role_name = vault_approle_auth_backend_role.remote_signer.role_name
}

resource "vault_approle_auth_backend_role_secret_id" "remote_signer" {
  backend   = vault_auth_backend.remote_signer.path
  role_name = vault_approle_auth_backend_role.remote_signer.role_name
  wrapping_ttl = 86400
}

output "role_id" {
  value = data.vault_approle_auth_backend_role_id.remote_signer.role_id
}

// DO NOT DO THIS! this is only for local tests. Pass the secret to app directly when provisioning
output "wrapped_secret_id" {
  value = nonsensitive(vault_approle_auth_backend_role_secret_id.remote_signer.wrapping_token)
}