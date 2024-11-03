#!/usr/bin/env bash

vault server -dev -dev-root-token-id="my-token" &

export VAULT_ADDR='http://localhost:8200'
vault login "my-token"
vault auth enable approle

vault policy write secretpaths - <<EOF
# allow read access to all policies
path "sys/policies/acl/*" {
  capabilities = ["read", "list"]
}
path "secret/*" {
  capabilities = ["list"]
}
EOF

vault write auth/approle/role/secretpaths \
    token_policies="secretpaths"

export APPROLE_ROLE_ID=$(vault read --field=data auth/approle/role/secretpaths/role-id --format=json | jq '.role_id' -r)
export APPROLE_SECRET_ID=$(vault write --field=data -f auth/approle/role/secretpaths/secret-id --format=json | jq '.secret_id' -r)

