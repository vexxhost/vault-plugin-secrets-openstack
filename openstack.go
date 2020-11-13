package mock

import (
	"net"
	"net/http"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/applicationcredentials"
	log "github.com/sirupsen/logrus"
	"opendev.org/vexxhost/openstack-operator/utils/tlsutils"
)

// authenticate in OpenStack and obtain service endpoint
func (b *backend) openstackClient(opts *gophercloud.AuthOptions, regionName string) (*gophercloud.ServiceClient, error) {

	providerClient, err := openstack.NewClient(opts.IdentityEndpoint)
	if err != nil {
		return nil, err
	}

	tlsConfig, err := tlsutils.CreateTLSConfig("OPENSTACK")
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
func (b *backend) CreateApplicationCredential(client *gophercloud.ServiceClient, userID string, name string) (string, string, error) {
	opts := applicationcredentials.CreateOpts{
		Name: name,
	}
	credential, err := applicationcredentials.Create(client, userID, opts).Extract()
	if err != nil {
		log.Errorf("Create applicationCredential failed - %s", err.Error())
		return "", "", err
	}
	return credential.ID, credential.secret, nil
}

// DeleteApplicationCredential deletes the applicationCredential
func (b *backend) DeleteApplicationCredential(client *gophercloud.ServiceClient, userID string, id string) error {
	if err := applicationcredentials.Delete(client, userID, id).Err; err != nil {
		log.Errorf("Delete applicationCredential failed - %s", err.Error())
	}
	return nil
}
