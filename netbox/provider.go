package netbox

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	runtimeclient "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/innovationnorway/go-netbox/plumbing"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"host": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NETBOX_HOST", nil),
			},

			"token": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("NETBOX_TOKEN", nil),
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"netbox_ipam_available_prefixes": dataSourceIpamAvailablePrefixes(),
			"netbox_ipam_prefix":             dataSourceIpamPrefix(),
			"netbox_ipam_prefixes":           dataSourceIpamPrefixes(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"netbox_ipam_available_prefix": resourceIpamAvailablePrefix(),
			"netbox_ipam_prefix":           resourceIpamPrefix(),
		},

		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	host := d.Get("host").(string)
	token := d.Get("token").(string)

	var diags diag.Diagnostics

	u, err := url.Parse(host)
	if err != nil {
		return nil, diag.Errorf("Unable to parse host: %s", err)
	}

	if u.Scheme == "" {
		u.Scheme = "https"
	}

	if u.Path == "" {
		u.Path = plumbing.DefaultBasePath
	}

	t := runtimeclient.New(u.Host, u.Path, []string{u.Scheme})
	t.Transport = logging.NewTransport("Netbox", t.Transport)

	if token != "" {
		t.DefaultAuthentication = runtimeclient.APIKeyAuth("Authorization", "header",
			fmt.Sprintf("Token %v", token))
	}

	return plumbing.New(t, strfmt.Default), diags
}
