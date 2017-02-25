// DO NOT EDIT. Generated by 'gorums' plugin for protoc-gen-go
// Source file to edit is: config_quorumcall_tmpl

package dev

import "golang.org/x/net/context"

// ReadQC invokes a ReadQC quorum call on configuration c
// and returns the result as a ReadQCReply.
func (c *Configuration) ReadQC(ctx context.Context, args *ReadReq) (*ReadQCReply, error) {
	return c.mgr.readQC(ctx, c, args)
}

// ReadQCCustomReturn invokes a ReadQCCustomReturn quorum call on configuration c
// and returns the result as a ReadQCCustomReturnReply.
func (c *Configuration) ReadQCCustomReturn(ctx context.Context, args *ReadReq) (*ReadQCCustomReturnReply, error) {
	return c.mgr.readQCCustomReturn(ctx, c, args)
}

// WriteQCPerNode invokes the WriteQCPerNode on each node in configuration c,
// with the argument returned by the provided perNodeArg function
// and returns the result as a WriteQCPerNodeReply.
func (c *Configuration) WriteQCPerNode(ctx context.Context, perNodeArg func(nodeID uint32) *Reply) (*WriteQCPerNodeReply, error) {
	return c.mgr.writeQCPerNode(ctx, c, perNodeArg)
}

// WriteQCWithReq invokes a WriteQCWithReq quorum call on configuration c
// and returns the result as a WriteQCWithReqReply.
func (c *Configuration) WriteQCWithReq(ctx context.Context, args *Reply) (*WriteQCWithReqReply, error) {
	return c.mgr.writeQCWithReq(ctx, c, args)
}
