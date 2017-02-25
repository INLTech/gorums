// DO NOT EDIT. Generated by 'gorums' plugin for protoc-gen-go
// Source file to edit is: node_tmpl

package dev

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
)

// Node encapsulates the state of a node on which a remote procedure call
// can be made.
type Node struct {
	// Only assigned at creation.
	id   uint32
	self bool
	addr string
	conn *grpc.ClientConn

	GorumsRPCClient GorumsRPCClient

	WriteMulticastClient GorumsRPC_WriteMulticastClient

	sync.Mutex
	lastErr error
	latency time.Duration
}

func (n *Node) connect(opts ...grpc.DialOption) error {
	var err error
	n.conn, err = grpc.Dial(n.addr, opts...)
	if err != nil {
		return fmt.Errorf("dialing node failed: %v", err)
	}

	n.GorumsRPCClient = NewGorumsRPCClient(n.conn)

	n.WriteMulticastClient, err = n.GorumsRPCClient.WriteMulticast(context.Background())
	if err != nil {
		return fmt.Errorf("stream creation failed: %v", err)
	}

	return nil
}

func (n *Node) close() error {
	// TODO: Log error, mainly care about the connection error below.
	// We should log this error, but we currently don't have access to the
	// logger in the manager.
	_, _ = n.WriteMulticastClient.CloseAndRecv()

	if err := n.conn.Close(); err != nil {
		return fmt.Errorf("conn close error: %v", err)
	}
	return nil
}
