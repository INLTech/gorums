// DO NOT EDIT. Generated by 'gorums' plugin for protoc-gen-go
// Source file to edit is: config_correctable_prelim_tmpl

package dev

import (
	"sync"

	"golang.org/x/net/context"
)

// ReadCorrectablePrelim is a reference to a correctable quorum call
// with server side preliminary reply support.
type ReadCorrectablePrelim struct {
	sync.Mutex
	reply    *ReadCorrectablePrelimReply
	level    int
	err      error
	done     bool
	watchers []*struct {
		level int
		ch    chan struct{}
	}
	donech chan struct{}
}

// ReadCorrectablePrelim asynchronously invokes a correctable ReadCorrectablePrelim quorum call
// with server side preliminary reply support on configuration c and returns a
// ReadCorrectablePrelim which can be used to inspect any replies or errors
// when available.
func (c *Configuration) ReadCorrectablePrelim(ctx context.Context, args *ReadReq) *ReadCorrectablePrelim {
	corr := &ReadCorrectablePrelim{
		level:  LevelNotSet,
		donech: make(chan struct{}),
	}
	go func() {
		c.mgr.readCorrectablePrelimCorrectablePrelim(ctx, c, corr, args)
	}()
	return corr
}

// Get returns the reply, level and any error associated with the
// ReadCorrectablePrelimCorrectablePremlim. The method does not block until a (possibly
// itermidiate) reply or error is available. Level is set to LevelNotSet if no
// reply has yet been received. The Done or Watch methods should be used to
// ensure that a reply is available.
func (c *ReadCorrectablePrelim) Get() (*ReadCorrectablePrelimReply, int, error) {
	c.Lock()
	defer c.Unlock()
	return c.reply, c.level, c.err
}

// Done returns a channel that's closed when the correctable ReadCorrectablePrelim
// quorum call is done. A call is considered done when the quorum function has
// signaled that a quorum of replies was received or that the call returned an
// error.
func (c *ReadCorrectablePrelim) Done() <-chan struct{} {
	return c.donech
}

// Watch returns a channel that's closed when a reply or error at or above the
// specified level is available. If the call is done, the channel is closed
// disregardless of the specified level.
func (c *ReadCorrectablePrelim) Watch(level int) <-chan struct{} {
	ch := make(chan struct{})
	c.Lock()
	if level < c.level {
		close(ch)
		c.Unlock()
		return ch
	}
	c.watchers = append(c.watchers, &struct {
		level int
		ch    chan struct{}
	}{level, ch})
	c.Unlock()
	return ch
}

func (c *ReadCorrectablePrelim) set(reply *ReadCorrectablePrelimReply, level int, err error, done bool) {
	c.Lock()
	if c.done {
		c.Unlock()
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
		c.Unlock()
		return
	}
	for i := range c.watchers {
		if c.watchers[i] != nil && c.watchers[i].level <= level {
			close(c.watchers[i].ch)
			c.watchers[i] = nil
		}
	}
	c.Unlock()
}
