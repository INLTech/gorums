// DO NOT EDIT. Generated by 'gorums' plugin for protoc-gen-go
// Source file to edit is: mgr_multicast_tmpl

package dev

import "golang.org/x/net/context"

func (m *Manager) writeMulticast(ctx context.Context, c *Configuration, args *Reply) error {
	for _, node := range c.nodes {
		go func(n *Node) {
			err := n.WriteMulticastClient.Send(args)
			if err == nil {
				return
			}
			if m.logger != nil {
				m.logger.Printf("%d: writeMulticast stream send error: %v", n.id, err)
			}
		}(node)
	}

	return nil
}
