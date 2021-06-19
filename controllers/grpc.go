package controllers

import (
	"context"
	"fmt"
	"strconv"

	gslbv1alpha1 "github.com/snapp-cab/gslb-controller/api/v1alpha1"
	"gitlab.snapp.ir/snappcloud/consul-gslb-driver/pkg/connection"
	"gitlab.snapp.ir/snappcloud/consul-gslb-driver/pkg/gslbi"
	"google.golang.org/grpc"
)

// Attacher implements create/delete operations against a remote gslbi driver.
type Client interface {
	CreateGslbcon(ctx context.Context, gslbcon *gslbv1alpha1.GslbContent) (id string, deleted bool, err error)
	DeleteGslbcon(ctx context.Context, gslbcon *gslbv1alpha1.GslbContent) error
}

type creater struct {
	conn *grpc.ClientConn
}

var grpcClient *creater

// NewClient provides a new Client object.
func NewClient(conn *grpc.ClientConn) {
	grpcClient = &creater{
		conn: conn,
	}
}

func (c *creater) CreateGslbcon(ctx context.Context, gslbcon *gslbv1alpha1.GslbContent) (id string, deleted bool, err error) {
	client := gslbi.NewControllerClient(c.conn)
	w, err := strconv.Atoi(gslbcon.Spec.Backend.Weight)
	if err != nil {
		return "", false, fmt.Errorf("invalid value for weight: %v, %w", gslbcon.Spec.Backend.Weight, err)
	}
	req := gslbi.CreateGSLBRequest{
		Name:        gslbcon.Name,
		ServiceName: gslbcon.Spec.ServiceName,
		Host:        gslbcon.Spec.Backend.Host,
		Weight:      int32(w),
		Parameters: map[string]string{
			"probe_timeout":  strconv.Itoa(int(gslbcon.Spec.Backend.Probe.TimeoutSeconds)),
			"probe_interval": strconv.Itoa(int(gslbcon.Spec.Backend.Probe.PeriodSeconds)),
			"probe_scheme":   gslbcon.Spec.Backend.Probe.HTTPGet.Scheme,
			"probe_address":  gslbcon.Spec.Backend.Probe.HTTPGet.Host,
		},
	}

	rsp, err := client.CreateGSLB(ctx, &req)
	if err != nil {
		return "", connection.IsFinalError(err), err
	}
	return rsp.Gslb.GslbId, false, nil
}

func (c *creater) DeleteGslbcon(ctx context.Context, gslbcon *gslbv1alpha1.GslbContent) error {
	client := gslbi.NewControllerClient(c.conn)

	req := gslbi.DeleteGSLBRequest{
		GslbId: gslbcon.Name,
	}

	_, err := client.DeleteGSLB(ctx, &req)
	return err
}
