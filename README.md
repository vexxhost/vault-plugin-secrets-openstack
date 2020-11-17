# vault-plugin-secrets-openstack
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