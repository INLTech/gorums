// DO NOT EDIT. Generated by 'gorums' plugin for protoc-gen-go
// Source file to edit is: config_correctable_tmpl

package dev

import (
	"sync"

	"golang.org/x/net/context"
)

// ReadCorrectable asynchronously invokes a
// correctable Read quorum call on configuration c and returns a
// ReadCorrectable which can be used to inspect any replies or errors
// when available.
func (c *Configuration) ReadCorrectable(ctx context.Context, args *ReadRequest) *ReadCorrectable {
	corr := &ReadCorrectable{
		level:  LevelNotSet,
		donech: make(chan struct{}),
	}
	go func() {
		c.mgr.readCorrectable(ctx, c, corr, args)
	}()
	return corr
}

// ReadCorrectable is a reference to a correctable Read quorum call.
type ReadCorrectable struct {
	mu       sync.Mutex
	reply    *ReadReply
	level    int
	err      error
	done     bool
	watchers []*struct {
		level int
		ch    chan struct{}
	}
	donech chan struct{}
}

// Get returns the reply, level and any error associated with the
// ReadCorrectable. The method does not block until a (possibly
// itermidiate) reply or error is available. Level is set to LevelNotSet if no
// reply has yet been received. The Done or Watch methods should be used to
// ensure that a reply is available.
func (c *ReadCorrectable) Get() (*ReadReply, int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.reply, c.level, c.err
}

// Done returns a channel that's closed when the correctable Read
// quorum call is done. A call is considered done when the quorum function has
// signaled that a quorum of replies was received or that the call returned an
// error.
func (c *ReadCorrectable) Done() <-chan struct{} {
	return c.donech
}

// Watch returns a channel that's closed when a reply or error at or above the
// specified level is available. If the call is done, the channel is closed
// disregardless of the specified level.
func (c *ReadCorrectable) Watch(level int) <-chan struct{} {
	ch := make(chan struct{})
	c.mu.Lock()
	if level < c.level {
		close(ch)
		c.mu.Unlock()
		return ch
	}
	c.watchers = append(c.watchers, &struct {
		level int
		ch    chan struct{}
	}{level, ch})
	c.mu.Unlock()
	return ch
}

func (c *ReadCorrectable) set(reply *ReadReply, level int, err error, done bool) {
	c.mu.Lock()
	if c.done {
		c.mu.Unlock()
		panic("set(...) called on a done correctable")
	}
	c.reply, c.level, c.err, c.done = reply, level, err, done
	if done {
		close(c.donech)
		for _, watcher := range c.watchers {
			if watcher != nil {
				close(watcher.ch)
			}
		}
		c.mu.Unlock()
		return
	}
	for i := range c.watchers {
		if c.watchers[i] != nil && c.watchers[i].level <= level {
			close(c.watchers[i].ch)
			c.watchers[i] = nil
		}
	}
	c.mu.Unlock()
}
