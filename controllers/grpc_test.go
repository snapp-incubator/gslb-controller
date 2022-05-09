package controllers

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"testing"

	"github.com/m-yosefpor/utils"
	gslbv1alpha1 "github.com/snapp-cab/gslb-controller/api/v1alpha1"
	"github.com/snapp-incubator/consul-gslb-driver/pkg/gslbi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func AtoiSyntaxError(str string) error {
	return &strconv.NumError{
		Func: "Atoi",
		Num:  str,
		Err:  strconv.ErrSyntax,
	}
}

type mockControllerServer struct {
	gslbi.UnimplementedControllerServer
}

func dialer() func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)

	server := grpc.NewServer()

	gslbi.RegisterControllerServer(server, &mockControllerServer{})

	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()

	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}

func TestControllerServerClient_CreateGslbcon(t *testing.T) {
	tests := []struct {
		name    string
		gslbCon *gslbv1alpha1.GslbContent
		err     error
	}{
		{
			"invalid request: non-convertable weight",
			&gslbv1alpha1.GslbContent{
				Spec: gslbv1alpha1.GslbContentSpec{
					Backend: gslbv1alpha1.Backend{
						Probe: gslbv1alpha1.Probe{
							Handler: gslbv1alpha1.Handler{
								HTTPGet: &gslbv1alpha1.HTTPGetAction{
									Host:   "google.com",
									Scheme: "http",
								},
							},
						},
						Weight: "a1",
					},
				},
			},
			fmt.Errorf("invalid value for weight: %v, %w", "a1", AtoiSyntaxError("a1")),
		},
		{
			"invalid request: empty weight",
			&gslbv1alpha1.GslbContent{
				Spec: gslbv1alpha1.GslbContentSpec{
					Backend: gslbv1alpha1.Backend{
						Probe: gslbv1alpha1.Probe{
							Handler: gslbv1alpha1.Handler{
								HTTPGet: &gslbv1alpha1.HTTPGetAction{
									Host:   "google.com",
									Scheme: "http",
								},
							},
						},
						Weight: "",
					},
				},
			},
			fmt.Errorf("invalid value for weight: %v, %w", "", AtoiSyntaxError("")),
		},
		{
			"invalid request: float weight",
			&gslbv1alpha1.GslbContent{
				Spec: gslbv1alpha1.GslbContentSpec{
					Backend: gslbv1alpha1.Backend{
						Probe: gslbv1alpha1.Probe{
							Handler: gslbv1alpha1.Handler{
								HTTPGet: &gslbv1alpha1.HTTPGetAction{
									Host:   "google.com",
									Scheme: "http",
								},
							},
						},
						Weight: "1.0",
					},
				},
			},
			fmt.Errorf("invalid value for weight: %v, %w", "1.0", AtoiSyntaxError("1.0")),
		},
		{
			"invalid request: nil HTTPGet",
			&gslbv1alpha1.GslbContent{
				Spec: gslbv1alpha1.GslbContentSpec{
					Backend: gslbv1alpha1.Backend{
						Weight: "1",
					},
				},
			},
			fmt.Errorf("Only HTTPGet is supported"),
		},
		{
			"valid request: convertable weight",
			&gslbv1alpha1.GslbContent{
				Spec: gslbv1alpha1.GslbContentSpec{
					Backend: gslbv1alpha1.Backend{
						Probe: gslbv1alpha1.Probe{
							Handler: gslbv1alpha1.Handler{
								HTTPGet: &gslbv1alpha1.HTTPGetAction{
									Host:   "google.com",
									Scheme: "http",
								},
							},
						},
						Weight: "3",
					},
				},
			},
			nil,
		},
		{
			"valid request: convertable weight, empty httpget",
			&gslbv1alpha1.GslbContent{
				Spec: gslbv1alpha1.GslbContentSpec{
					Backend: gslbv1alpha1.Backend{
						Probe: gslbv1alpha1.Probe{
							Handler: gslbv1alpha1.Handler{
								HTTPGet: &gslbv1alpha1.HTTPGetAction{},
							},
						},
						Weight: "2",
					},
				},
			},
			nil,
		},
	}

	ctx := context.Background()

	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &creater{
				conn: conn,
			}
			_, _, err := client.CreateGslbcon(context.Background(), tt.gslbCon)

			if err != nil && !utils.IsErrorMessage(err, tt.err) {
				t.Error("error: expected", tt.err, "received", err)
			}
		})
	}
}

func (cs *mockControllerServer) CreateGSLB(ctx context.Context, req *gslbi.CreateGSLBRequest) (*gslbi.CreateGSLBResponse, error) {
	resp := &gslbi.CreateGSLBResponse{
		Gslb: &gslbi.Gslb{
			GslbId: "someid", //tod

		},
	}
	return resp, nil
}

func TestControllerServerClient_DeleteGslbcon(t *testing.T) {
	tests := []struct {
		name    string
		gslbCon *gslbv1alpha1.GslbContent
		err     error
	}{
		{
			"valid request: convertable weight",
			&gslbv1alpha1.GslbContent{
				Spec: gslbv1alpha1.GslbContentSpec{
					Backend: gslbv1alpha1.Backend{
						Probe: gslbv1alpha1.Probe{
							Handler: gslbv1alpha1.Handler{
								HTTPGet: &gslbv1alpha1.HTTPGetAction{
									Host:   "google.com",
									Scheme: "http",
								},
							},
						},
						Weight: "3",
					},
				},
			},
			nil,
		},
	}

	ctx := context.Background()

	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &creater{
				conn: conn,
			}
			err := client.DeleteGslbcon(context.Background(), tt.gslbCon)

			if err != nil && !utils.IsErrorMessage(err, tt.err) {
				t.Error("error: expected", tt.err, "received", err)
			}
		})
	}
}

func (cs *mockControllerServer) DeleteGSLB(ctx context.Context, req *gslbi.DeleteGSLBRequest) (*gslbi.DeleteGSLBResponse, error) {
	return &gslbi.DeleteGSLBResponse{}, nil
}
