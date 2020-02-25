package grpc

import (
	"context"

	"github.com/ParadigmFoundation/go-obm/grpc/types"
	"google.golang.org/grpc"
)

type Client struct {
	types.OrderBookManagerClient
}

func NewClient(ctx context.Context, addr string) (*Client, error) {
	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return &Client{
		types.NewOrderBookManagerClient(conn),
	}, nil
}
