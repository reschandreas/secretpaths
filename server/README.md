# server component of secretpaths

To be able to read all paths and policies, we need the following policy:
```bash
vault policy write secretpaths - <<EOF
# allow read access to all policies
path "sys/policies/acl/*" {
  capabilities = ["read", "list"]
}
path "secret/*" {
  capabilities = ["list"]
}
EOF
```

# Using approles

Make sure to enable the approle auth method in vault.

```bash
vault write auth/approle/role/secretpaths \
    token_policies="secretpaths"
```

```bash
vault read auth/approle/role/secretpaths/role-id
vault write -f auth/approle/role/secretpaths/secret-id
```

```bash

## Setup the dev environment
```bash
export VAULT_ADDR='http://localhost:8200'
vault login "my-token"
vault auth enable approle

```