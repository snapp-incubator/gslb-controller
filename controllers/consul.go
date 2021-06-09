package controllers

import (
	"context"
	"fmt"
	"time"

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
	catsvcList, _, err := consul.cat.Service(gslb.Name, "", &consulapi.QueryOptions{})
	if err != nil {
		return gslb, fmt.Errorf("failed to query consul for servic: %w", err)
	}
	if len(catsvcList) == 0 {
		return gslb, fmt.Errorf("NotFound")
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
	// TODO check if serviceName is not used, else raise error
	for _, b := range gslb.Spec.Backends {
		var probeHost string
		if b.Probe.HTTPGet.Host == "" {
			probeHost = b.Host
		} else {
			probeHost = b.Probe.HTTPGet.Host
		}
		header := make(map[string][]string)
		for _, h := range b.Probe.HTTPGet.HTTPHeaders {
			header[h.Name] = []string{h.Value}
		}
		reg := &consulapi.CatalogRegistration{
			Node:    gslb.Namespace + "-" + gslb.Name + "-" + b.Name,
			Address: b.Host,
			NodeMeta: map[string]string{
				"external-node":  "true",
				"external-probe": "false",
			},
			Service: &consulapi.AgentService{
				ID:      gslb.Spec.ServiceName,
				Service: gslb.Spec.ServiceName,
			},
			Checks: consulapi.HealthChecks{
				&consulapi.HealthCheck{
					Name: "http-check",
					Definition: consulapi.HealthCheckDefinition{
						HTTP:             probeHost,
						IntervalDuration: time.Duration(b.Probe.PeriodSeconds),
						TimeoutDuration:  time.Duration(b.Probe.TimeoutSeconds),
						Header:           header,
					},
				},
			},
		}
		_, err := consul.cat.Register(reg, &consulapi.WriteOptions{})
		if err != nil {
			return fmt.Errorf("failed to create consul service: %w", err)
		}
	}

	return nil
}
