# OpenStack Secret Engine for Hashicorp Vault

This is a Vault secret engine plugin which allows you to generate OpenStack
application credentials which can automatically expire (and also scoped out
to specific roles as well).

## Usage

Move the compiled plugin into Vault's configured [`plugin_directory`](https://www.vaultproject.io/docs/configuration/index.html#plugin_directory):

```shell
mv vault-plugin-secrets-openstack /etc/vault/plugins/vault-plugin-secrets-openstack
```

Calculate the SHA256 of the plugin and register it in Vault's plugin catalog:

```shell
export SHA256=$(shasum -a 256 "/etc/vault/plugins/vault-plugin-secrets-openstack" | cut -d' ' -f1)

vault write sys/plugins/catalog/vault-plugin-secrets-openstack \
              sha_256="${SHA256}" \
              command="vault-plugin-secrets-openstack"

Success! Data written to: sys/plugins/catalog/vault-plugin-secrets-openstack
```

### Configuration

Enable and configure the secret engine:

```shell
vault secrets enable -path="openstack" -plugin-name="vault-plugin-secrets-openstack" plugin
vault write openstack/config/lease ttl=60
vault write openstack/config/auth auth_url="https://auth.vexxhost.net/v3" \
                                    user_id="<user_id>" \
                                    password="<password>"
```

The example above configures a default lease of 60 seconds and points to the
VEXXHOST public cloud authentication endpoint.

#### Authentication Options

The plugin supports two authentication methods:

**Username/Password** (recommended for multi-project support):
- `auth_url` - OpenStack authentication URL
- `user_id` or `username` - User credentials
- `password` - User password
- `user_domain_id` / `user_domain_name` - Domain for user authentication

**Application Credentials** (for single-project use):
- `auth_url` - OpenStack authentication URL
- `application_credential_id` or `application_credential_name` - Application credential identifier
- `application_credential_secret` - Application credential secret

**Additional Options:**
- `region_name` - Region name for endpoint selection
- `cacert` - PEM-encoded CA certificate for TLS verification
- `cert` / `key` - PEM-encoded client certificate and key for mutual TLS
- `insecure` - Skip TLS verification (not recommended for production)

### Rolesets

Create a roleset to define what application credentials will be created. Rolesets
can optionally specify a project scope.

**Single-project mode** (with application credential auth):

```shell
vault write openstack/roleset/member roles=-<<EOF
[
  {
    "id": "9fe2ff9ee4384b1894a90878d3e92bab"
  }
]
EOF
```

**Multi-project mode** (with username/password auth):

```shell
# Roleset for project A
vault write openstack/roleset/project-a-member \
    project_id="<project_a_id>" \
    roles=-<<EOF
[
  {
    "id": "9fe2ff9ee4384b1894a90878d3e92bab"
  }
]
EOF

# Roleset for project B
vault write openstack/roleset/project-b-admin \
    project_id="<project_b_id>" \
    roles=-<<EOF
[
  {
    "id": "admin_role_id"
  }
]
EOF
```

Roleset options:
- `project_id` / `project_name` - Project to scope the application credential to
- `project_domain_id` / `project_domain_name` - Domain for project scoping
- `roles` - JSON array of roles for the application credential

> **Note:** When using application credential authentication, project fields in
> rolesets are not supported (application credentials are bound to their original
> project). Use username/password authentication for multi-project support.

### Generating Credentials

To create an application credential which will expire within 60 seconds based
on the configured time to live:

```shell
vault read openstack/creds/member
Key                              Value
---                              -----
lease_id                         openstack/creds/member/alWy2bskdhoroBKSUlKX6UgR
lease_duration                   1m
lease_renewable                  false
application_credential_id        <snip>
application_credential_secret    <snip>
```

You'll see that an application credential was issued once you run this command:

```shell
openstack application credential list
+----------------------------------+----------------------------------------+----------------------------------+------------------------------------------+----------------------------+
| ID                               | Name                                   | Project ID                       | Description                              | Expires At                 |
+----------------------------------+----------------------------------------+----------------------------------+------------------------------------------+----------------------------+
| bc70b161405740b9927d10b45be7502c | vault-member-token-1674140730969888877 | 8709ca2640344a4ba85cba0a1d6eea69 | Created by Vault at 2023-01-19T15:05:30Z | 2023-01-19T15:15:30.970151 |
+----------------------------------+----------------------------------------+----------------------------------+------------------------------------------+----------------------------+
```

After the 60 seconds are up, you'll see that the token no longer exists in there
and the lease is revoked.

## Development

In order to run the plugin locally, you'll need to have Vault installed inside
your `$PATH` and run the following inside your terminal which will both build
the plugin and start up Vault.

```shell
make start
```

At this point, you can refer to the [Usage](#Usage) section for how to enable the plugin
and interact with it.  The one thing to note is that you'll have to make sure
that you export the `VAULT_ADDR` in your interactions with the tool.
