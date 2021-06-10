package controllers

import (
	"context"
	"fmt"
	"strconv"
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
	catsvcList, _, err := consul.cat.Service(gslb.Spec.ServiceName, "", &consulapi.QueryOptions{})
	if err != nil {
		return gslb, fmt.Errorf("failed to query consul for service: %w", err)
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
		var healthcheck consulapi.HealthCheck
		switch {
		case b.Probe.HTTPGet != nil:
			healthcheck = httpHealthcheck(ctx, b)
		case b.Probe.TCPSocket != nil:
			healthcheck = tcpHealthcheck(ctx, b)
		case b.Probe.Exec != nil:
			healthcheck = execHealthcheck(ctx, b)
		default:
			return fmt.Errorf("at least one check type must be specified: [\"httpGet\",\"exec\",\"tcpSocket\"]")
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
				&healthcheck,
			},
		}
		_, err := consul.cat.Register(reg, &consulapi.WriteOptions{})
		if err != nil {
			return fmt.Errorf("failed to register consul service: %w", err)
		}
	}

	return nil
}

func DeleteGslb(ctx context.Context, gslb *gslbv1alpha1.Gslb) error {
	for _, b := range gslb.Spec.Backends {
		dereg := &consulapi.CatalogDeregistration{
			Node: gslb.Namespace + "-" + gslb.Name + "-" + b.Name,
		}
		_, err := consul.cat.Deregister(dereg, &consulapi.WriteOptions{})
		if err != nil {
			return fmt.Errorf("failed to deregister consul service: %w", err)
		}
	}
	return nil
}

func httpHealthcheck(ctx context.Context, b gslbv1alpha1.Backend) consulapi.HealthCheck {
	var probeHost string
	if b.Probe.HTTPGet.Host == "" {
		probeHost = b.Host
	} else {
		probeHost = b.Probe.HTTPGet.Host
	}
	if b.Probe.HTTPGet.Port != 0 {
		probeHost = probeHost + ":" + strconv.Itoa(int(b.Probe.HTTPGet.Port))
	}
	header := make(map[string][]string)
	for _, h := range b.Probe.HTTPGet.HTTPHeaders {
		header[h.Name] = []string{h.Value}
	}
	return consulapi.HealthCheck{
		Name:   "http-check",
		Status: "passing",
		Definition: consulapi.HealthCheckDefinition{
			HTTP:             b.Probe.HTTPGet.Scheme + "://" + probeHost + b.Probe.HTTPGet.Path,
			IntervalDuration: time.Duration(b.Probe.PeriodSeconds) * time.Second,
			TimeoutDuration:  time.Duration(b.Probe.TimeoutSeconds) * time.Second,
			Header:           header,
		},
	}
}

func tcpHealthcheck(ctx context.Context, b gslbv1alpha1.Backend) consulapi.HealthCheck {
	return consulapi.HealthCheck{}
}

func execHealthcheck(ctx context.Context, b gslbv1alpha1.Backend) consulapi.HealthCheck {
	return consulapi.HealthCheck{}
}
