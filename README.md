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

In order to get started, you'll need to enable and configure the secret engine
by running the following:

```shell
vault secrets enable -path=openstack -plugin-name=vault-plugin-secrets-openstack
vault write openstack/config/lease ttl=60
vault write openstack/config/auth IdentityEndpoint="https://auth.vexxhost.net/v3" \
                                    UserID="<user_id>" \
                                    Password="<password>" \
                                    TenantID="<tenant_id>"
```

The example above configures a default lease of 60 seconds and points to the
VEXXHOST public cloud authentication endpoint.  You'll need to replace your
user ID, password and tenant ID in the example above.  There are some other
configuration options that you can use which you can lookup using `vault read
openstack/config/auth` command.

The next step you'll need to do is create a roleset which will be used when
creating application credentials.  In this example, it is using the default
member role inside the VEXXHOST public cloud.

```shell
vault write openstack/roleset/member Roles=-<<EOF
[
  {
    "ID": "9fe2ff9ee4384b1894a90878d3e92bab"
  }
]
EOF
```

Now, in order to create an application credential which will expire within
60 seconds based on the configured time to live, you can run a `read` on the
`auth` endpoint.

```shell
vault read openstack/creds/member
ey                              Value
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
