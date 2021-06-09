package controllers

import (
	"context"
	"fmt"

	consulapi "github.com/hashicorp/consul/api"
	gslbv1alpha1 "github.com/snapp-cab/gslb-controller/api/v1alpha1"
)

type Consul struct {
	c   *consulapi.Client
	cat *consulapi.Catalog
}

var consul Consul

func init() {
	config := &consulapi.Config{
		Address:    "consul.apps.private.okd4.teh-1.snappcloud.io",
		Scheme:     "http",
		Datacenter: "teh1",
	}
	client, err := consulapi.NewClient(config)
	if err != nil {
		panic(err)
	}
	consul.c = client
	consul.cat = client.Catalog()
}

func GetGslb(ctx context.Context, gslb *gslbv1alpha1.Gslb) (*gslbv1alpha1.Gslb, error) {
	_, _, err := consul.cat.Service(gslb.Name, "", &consulapi.QueryOptions{})
	if err != nil {
		return gslb, fmt.Errorf("failed to query consul for servic: %w", err)
	}
	return gslb, nil
}

func CreateGslb(ctx context.Context, gslb *gslbv1alpha1.Gslb) error {
	return CreateOrUpdateGslb(ctx, gslb)
}

func UpdateGslb(ctx context.Context, gslb *gslbv1alpha1.Gslb) error {
	return CreateOrUpdateGslb(ctx, gslb)
}

func CreateOrUpdateGslb(ctx context.Context, gslb *gslbv1alpha1.Gslb) error {
	reg := &consulapi.CatalogRegistration{
		Node:    gslb.Namespace + "-" + gslb.Name,
		Address: gslb.Spec.Host,
		NodeMeta: map[string]string{
			"external-node":  "true",
			"external-probe": "false",
		},
		Service: &consulapi.AgentService{
			ID:      gslb.Name,
			Service: gslb.Name,
		},
		// Checks: consulapi.HealthChecks{},
	}
	_, err := consul.cat.Register(reg, &consulapi.WriteOptions{})
	if err != nil {
		return fmt.Errorf("failed to create consul service: %w", err)
	}
	return nil
}
