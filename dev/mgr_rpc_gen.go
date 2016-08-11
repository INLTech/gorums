package dev

import (
	"time"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type readReply struct {
	nid   uint32
	reply *State
	err   error
}

func (m *Manager) read(c *Configuration, args *ReadRequest) (*ReadReply, error) {
	var (
		replyChan   = make(chan *readReply, c.n)
		ctx, cancel = context.WithCancel(context.Background())
	)

	callGRPC := func(node *Node) {
		reply := new(State)
		start := time.Now()
		err := grpc.Invoke(
			ctx,
			"/dev.Register/Read",
			args,
			reply,
			node.conn,
		)
		switch grpc.Code(err) { // nil -> codes.OK
		case codes.OK, codes.Canceled:
			node.setLatency(time.Since(start))
		default:
			node.setLastErr(err)
		}
		replyChan <- &readReply{node.id, reply, err}
	}

	if len(c.nodes) == 1 {
		// no need to create goroutine for calls on single node configurations
		callGRPC(c.nodes[0])
	} else {
		for _, n := range c.nodes {
			go callGRPC(n)
		}
	}

	var (
		replyValues = make([]*State, 0, c.n)
		reply       = &ReadReply{NodeIDs: make([]uint32, 0, c.n)}
		errCount    int
		quorum      bool
	)

	/*
		Alternative for time.After in select below: stop rpc timeout timer explicitly.

		See
		https://github.com/kubernetes/kubernetes/pull/23210/commits/e4b369e1d74ac8f2d2a20afce92d93c804afa5d2
		and
		https://github.com/golang/go/issues/8898l

		t := time.NewTimer(c.timeout)
		defer t.Stop()

		and change the corresponding select case below:

		case <-t.C:

		Actually gaven an +1% on the local read benchmark, so not implemted yet.
	*/

	for {

		select {
		case r := <-replyChan:
			if r.err != nil {
				errCount++
				goto terminationCheck
			}
			replyValues = append(replyValues, r.reply)
			reply.NodeIDs = append(reply.NodeIDs, r.nid)
			if reply.Reply, quorum = c.qspec.ReadQF(replyValues); quorum {
				cancel()
				return reply, nil
			}
		case <-time.After(c.timeout):
			cancel()
			return reply, TimeoutRPCError{c.timeout, errCount, len(replyValues)}
		}

	terminationCheck:
		if errCount+len(replyValues) == c.n {
			cancel()
			return reply, IncompleteRPCError{errCount, len(replyValues)}
		}

	}
}

type writeReply struct {
	nid   uint32
	reply *WriteResponse
	err   error
}

func (m *Manager) write(c *Configuration, args *State) (*WriteReply, error) {
	var (
		replyChan   = make(chan writeReply, c.n)
		ctx, cancel = context.WithCancel(context.Background())
	)

	for _, n := range c.nodes {
		go func(node *Node) {
			reply := new(WriteResponse)
			start := time.Now()
			err := grpc.Invoke(
				ctx,
				"/dev.Register/Write",
				args,
				reply,
				node.conn,
			)
			switch grpc.Code(err) { // nil -> codes.OK
			case codes.OK, codes.Canceled:
				node.setLatency(time.Since(start))
			default:
				node.setLastErr(err)
			}
			replyChan <- writeReply{node.id, reply, err}
		}(n)
	}

	var (
		replyValues = make([]*WriteResponse, 0, c.n)
		reply       = &WriteReply{NodeIDs: make([]uint32, 0, c.n)}
		errCount    int
		quorum      bool
	)

	for {

		select {
		case r := <-replyChan:
			if r.err != nil {
				errCount++
				goto terminationCheck
			}
			replyValues = append(replyValues, r.reply)
			reply.NodeIDs = append(reply.NodeIDs, r.nid)
			if reply.Reply, quorum = c.qspec.WriteQF(replyValues); quorum {
				cancel()
				return reply, nil
			}
		case <-time.After(c.timeout):
			cancel()
			return reply, TimeoutRPCError{c.timeout, errCount, len(replyValues)}
		}

	terminationCheck:
		if errCount+len(replyValues) == c.n {
			cancel()
			return reply, IncompleteRPCError{errCount, len(replyValues)}
		}
	}
}

func (m *Manager) writeAsync(c *Configuration, args *State) error {
	for _, node := range c.nodes {
		go func(n *Node) {
			err := n.writeAsyncClient.Send(args)
			if err == nil {
				return
			}
			if m.logger != nil {
				m.logger.Printf("%d: writeAsync stream send error: %v", n.id, err)
			}
		}(node)
	}

	return nil
}
