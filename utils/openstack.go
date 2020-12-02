package utils

import (
	"net"
	"net/http"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/applicationcredentials"
	log "github.com/sirupsen/logrus"
)

// authenticate in OpenStack and obtain service endpoint
func OpenstackClient(opts gophercloud.AuthOptions, regionName string) (*gophercloud.ServiceClient, error) {

	providerClient, err := openstack.NewClient(opts.IdentityEndpoint)
	if err != nil {
		return nil, err
	}

	tlsConfig, err := CreateTLSConfig("OPENSTACK")
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig:       tlsConfig,
	}
	providerClient.HTTPClient.Transport = transport

	if err = openstack.Authenticate(providerClient, opts); err != nil {
		return nil, err
	}

	eo := gophercloud.EndpointOpts{
		Region: regionName,
	}

	client, err := openstack.NewIdentityV3(providerClient, eo)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// CreateApplicationCredential creates a applicationCredential
func CreateApplicationCredential(client *gophercloud.ServiceClient, userID string, name string, roles []applicationcredentials.Role, ttl time.Duration) (string, string, error) { //, accessrules []applicationcredentials.AccessRule

	expireTime := time.Now().Add(time.Second * ttl)
	opts := applicationcredentials.CreateOpts{
		Name:      name,
		Roles:     roles,
		ExpiresAt: &expireTime,
		// AccessRules: accessrules,
	}
	credential, err := applicationcredentials.Create(client, userID, opts).Extract()
	if err != nil {
		log.Errorf("Create applicationCredential failed - %s", err.Error())
		return "", "", err
	}
	return credential.ID, credential.Secret, nil
}

// DeleteApplicationCredential deletes the applicationCredential
func DeleteApplicationCredential(client *gophercloud.ServiceClient, userID string, id string) error {
	if err := applicationcredentials.Delete(client, userID, id).Err; err != nil {
		log.Errorf("Delete applicationCredential failed - %s", err.Error())
	}
	return nil
}
