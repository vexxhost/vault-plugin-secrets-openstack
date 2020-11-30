# OpenStack Secret Engine for Hashicorp Vault

This is a Vault secret engine plugin which allows you to generate OpenStack
application credentials which can automatically expire (and also scoped out
to specific roles as well).

# Usage
In order to get started, you'll need to enable and configure the secret engine
by running the following:

```console
$ vault secrets enable -path=openstack vault-plugin-secrets-openstack
$ vault write openstack/config/lease ttl=60
$ vault write openstack/config/auth IdentityEndpoint="https://auth.vexxhost.net/v3" \
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

```console
$ vault write openstack/roleset/member Roles=-<<EOF
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

```console
$ vault read openstack/creds/member
ey                              Value
---                              -----
lease_id                         openstack/creds/member/alWy2bskdhoroBKSUlKX6UgR
lease_duration                   1m
lease_renewable                  true
application_credential_id        <snip>
application_credential_secret    <snip>
```

You'll see that an application credential was issued once you run this command:

```console
$ openstack application credential list
+----------------------------------+----------------------------------------+----------------------------------+-------------+------------+
| ID                               | Name                                   | Project ID                       | Description | Expires At |
+----------------------------------+----------------------------------------+----------------------------------+-------------+------------+
| bc70b161405740b9927d10b45be7502c | vault-member-token-1606779565965295355 | 8709ca2640344a4ba85cba0a1d6eea69 | None        | None       |
+----------------------------------+----------------------------------------+----------------------------------+-------------+------------+
```

After the 60 seconds are up, you'll see that the token no longer exists in there
and the lease is revoked.

## Development
In order to run the plugin locally, you'll need to have Vault installed inside
your `$PATH` and run the following inside your terminal which will both build
the plugin and start up Vault.

```console
$ make start
```

At this point, you can refer to the [Usage](#Usage) section for how to enable the plugin
and interact with it.  The one thing to note is that you'll have to make sure
that you export the `VAULT_ADDR` in your interactions with the tool.





Once the server is up and running, you'll have to export the correct
environment variable and enable the plug-in within Vault.

```console
$ export VAULT_ADDR='http://127.0.0.1:8200'
$ make enable
```

At this point, you can interact and use the secret engine which will be mounted
at the path `openstack/` by default.  If you make changes to the code, you will
need to restart all the processes.

Sample commands:
```bash

make enable

vault write openstack-secrets/config/lease ttl=300

vault read openstack-secrets/config/lease

vault write openstack-secrets/config/auth \
                    IdentityEndpoint="xxxxx" \
                    UserID="xxxxx" \
                    Password="xxxx" \
                    TenantID="xxxx" \
                    TenantName="xxxx" \
                    Region="xxxx"

vault read openstack-secrets/config/auth

vault write openstack-secrets/roleset/test Roles=-<<EOF
[
  {
    "ID": "9bb4c2b3d60742609673c01c7e749a3c"
  }
]
EOF

vault read openstack-secrets/roleset/test

vault read openstack-secrets/creds/test

```