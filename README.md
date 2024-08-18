# secretpaths

secretpaths is a tool for visualizing secrets and paths managed by policies in HashiCorp Vault and OpenBao. 
It provides an intuitive graphical interface to help engineers understand and manage policy-bound secrets and 
their access paths, enhancing security and transparency in secret management.

![show_paths_ui.png](./docs/images/show_paths_ui.png)


## Configuration

We use environment variables to configure the application. The following environment variables are available:

| Environment Variable | Description                                                                       | Default Value           |
|----------------------|-----------------------------------------------------------------------------------|-------------------------|
| `VAULT_ADDR`         | The address of the Vault server                                                   | `http://127.0.0.1:8200` |
| `VAULT_TOKEN`        | The token to authenticate with the Vault server, should NOT be used in production |                         |
| `VAULT_ROLE_ID`      | The role ID to authenticate with the Vault server                                 |                         |
| `VAULT_SECRET_ID`    | The secret ID to authenticate with the Vault server                               |                         |
| `KUBERNETES_ROLE`    | The role to authenticate with the Kubernetes server                               |                         |
| `VAULT_KV_ENGINE`    | The key-value engine to use in Vault                                              | `secret`                |