// DO NOT EDIT. Generated by github.com/relab/gorums/cmd/gentemplates
// Template source files to edit is in the 'dev' folder.

package gorums

const config_qc_tmpl = `
{{/* Remember to run 'make goldenanddev' after editing this file. */}}

{{- if not .IgnoreImports}}
package {{.PackageName}}

import (
	"fmt"
	"sync"

	"golang.org/x/net/context"
)

{{- end}}

{{range $elm := .Services}}

{{if .Multicast}}

// {{.MethodName}} is a one-way multicast operation, where args is sent to
// every node in configuration c. The call is asynchronous and has no response
// return value.
func (c *Configuration) {{.MethodName}}(ctx context.Context, args *{{.FQReqName}}) error {
	return c.mgr.{{.UnexportedMethodName}}(ctx, c, args)
}

{{- end -}}

{{if or (.QuorumCall) (.Future) (.Correctable)}}

// {{.TypeName}} encapsulates the reply from a {{.MethodName}} quorum call.
// It contains the id of each node of the quorum that replied and a single reply.
type {{.TypeName}} struct {
	NodeIDs []uint32
	*{{.FQRespName}}
}

func (r {{.TypeName}}) String() string {
	return fmt.Sprintf("node ids: %v | answer: %v", r.NodeIDs, r.{{.RespName}})
}
{{- end -}}

{{if and (not (.PerNodeArg)) (.QuorumCall)}}
// {{.MethodName}} invokes a {{.MethodName}} quorum call on configuration c
// and returns the result as a {{.TypeName}}.
func (c *Configuration) {{.MethodName}}(ctx context.Context, args *{{.FQReqName}}) (*{{.TypeName}}, error) {
	return c.mgr.{{.UnexportedMethodName}}(ctx, c, args)
}
{{- end -}}

{{if and (.PerNodeArg) (.QuorumCall)}}
// {{.MethodName}} invokes the {{.MethodName}} on each node in configuration c,
// with the argument returned by the provided perNodeArg function
// and returns the result as a {{.TypeName}}.
func (c *Configuration) {{.MethodName}}(ctx context.Context, {{.MethodArg}}) (*{{.TypeName}}, error) {
	return c.mgr.{{.UnexportedMethodName}}(ctx, c, {{.MethodArgUse}})
}
{{- end -}}

{{if .Future}}

// {{.MethodName}}Future is a reference to an asynchronous {{.MethodName}} quorum call invocation.
type {{.MethodName}}Future struct {
	reply *{{.TypeName}}
	err   error
	c     chan struct{}
}

// {{.MethodName}}Future asynchronously invokes a {{.MethodName}} quorum call
// on configuration c and returns a {{.MethodName}}Future which can be used to
// inspect the quorum call reply and error when available.
func (c *Configuration) {{.MethodName}}Future(ctx context.Context, args *{{.FQReqName}}) *{{.MethodName}}Future {
	f := new({{.MethodName}}Future)
	f.c = make(chan struct{}, 1)
	go func() {
		defer close(f.c)
		f.reply, f.err = c.mgr.{{.UnexportedMethodName}}(ctx, c, args)
	}()
	return f
}

// Get returns the reply and any error associated with the {{.MethodName}}Future.
// The method blocks until a reply or error is available.
func (f *{{.MethodName}}Future) Get() (*{{.TypeName}}, error) {
	<-f.c
	return f.reply, f.err
}

// Done reports if a reply and/or error is available for the {{.MethodName}}Future.
func (f *{{.MethodName}}Future) Done() bool {
	select {
	case <-f.c:
		return true
	default:
		return false
	}
}

{{- end -}}

{{if .Correctable}}

// {{.MethodName}}Correctable asynchronously invokes a
// correctable {{.MethodName}} quorum call on configuration c and returns a
// {{.MethodName}}Correctable which can be used to inspect any repies or errors
// when available.
func (c *Configuration) {{.MethodName}}Correctable(ctx context.Context, args *ReadRequest) *{{.MethodName}}Correctable {
	corr := &{{.MethodName}}Correctable{
		level:  LevelNotSet,
		donech: make(chan struct{}),
	}
	go func() {
		c.mgr.{{.UnexportedMethodName}}Correctable(ctx, c, corr, args)
	}()
	return corr
}

// {{.MethodName}}Correctable is a reference to a correctable {{.MethodName}} quorum call.
type {{.MethodName}}Correctable struct {
	mu       sync.Mutex
	reply    *{{.TypeName}}
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
// {{.MethodName}}Correctable. The method does not block until a (possibly
// itermidiate) reply or error is available. Level is set to LevelNotSet if no
// reply has yet been received. The Done or Watch methods should be used to
// ensure that a reply is available.
func (c *{{.MethodName}}Correctable) Get() (*{{.TypeName}}, int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.reply, c.level, c.err
}

// Done returns a channel that's closed when the correctable {{.MethodName}}
// quorum call is done. A call is considered done when the quorum function has
// signaled that a quorum of replies was received or that the call returned an
// error.
func (c *{{.MethodName}}Correctable) Done() <-chan struct{} {
	return c.donech
}

// Watch returns a channel that's closed when a reply or error at or above the
// specified level is available. If the call is done, the channel is closed
// disregardless of the specified level.
func (c *{{.MethodName}}Correctable) Watch(level int) <-chan struct{} {
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

func (c *{{.MethodName}}Correctable) set(reply *{{.TypeName}}, level int, err error, done bool) {
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

{{- end -}}

{{if .CorrectablePrelim}}

// {{.TypeName}} encapsulates the reply from a correctable {{.MethodName}} quorum call.
// It contains the id of each node of the quorum that replied and a single reply.
type {{.TypeName}} struct {
	NodeIDs []uint32
	*{{.FQRespName}}
}

func (r {{.TypeName}}) String() string {
	return fmt.Sprintf("node ids: %v | answer: %v", r.NodeIDs, r.{{.RespName}})
}

// {{.MethodName}}CorrectablePrelim asynchronously invokes a correctable {{.MethodName}} quorum call
// with server side preliminary reply support on configuration c and returns a
// {{.MethodName}}CorrectablePrelim which can be used to inspect any repies or errors
// when available.
func (c *Configuration) {{.MethodName}}CorrectablePrelim(ctx context.Context, args *{{.FQReqName}}) *{{.MethodName}}CorrectablePrelim {
	corr := &{{.MethodName}}CorrectablePrelim{
		level:  LevelNotSet,
		donech: make(chan struct{}),
	}
	go func() {
		c.mgr.{{.UnexportedMethodName}}CorrectablePrelim(ctx, c, corr, args)
	}()
	return corr
}

// {{.MethodName}}CorrectablePrelim is a reference to a correctable Read quorum call
// with server side preliminary reply support.
type {{.MethodName}}CorrectablePrelim struct {
	mu       sync.Mutex
	reply    *{{.TypeName}}
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
// {{.MethodName}}CorrectablePremlim. The method does not block until a (possibly
// itermidiate) reply or error is available. Level is set to LevelNotSet if no
// reply has yet been received. The Done or Watch methods should be used to
// ensure that a reply is available.
func (c *{{.MethodName}}CorrectablePrelim) Get() (*{{.TypeName}}, int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.reply, c.level, c.err
}

// Done returns a channel that's closed when the correctable {{.MethodName}}
// quorum call is done. A call is considered done when the quorum function has
// signaled that a quorum of replies was received or that the call returned an
// error.
func (c *{{.MethodName}}CorrectablePrelim) Done() <-chan struct{} {
	return c.donech
}

// Watch returns a channel that's closed when a reply or error at or above the
// specified level is available. If the call is done, the channel is closed
// disregardless of the specified level.
func (c *{{.MethodName}}CorrectablePrelim) Watch(level int) <-chan struct{} {
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

func (c *{{.MethodName}}CorrectablePrelim) set(reply *{{.TypeName}}, level int, err error, done bool) {
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

{{- end -}}

{{- end -}}
`

const mgr_correctable_tmpl = `
{{/* Remember to run 'make goldenanddev' after editing this file. */}}

{{$pkgName := .PackageName}}

{{if not .IgnoreImports}}
package {{$pkgName}}

import "golang.org/x/net/context"

{{end}}

{{range $elm := .Services}}

{{if .Correctable}}

func (m *Manager) {{.UnexportedMethodName}}Correctable(ctx context.Context, c *Configuration, corr *{{.MethodName}}Correctable, args *{{.FQReqName}}) {
	replyChan := make(chan {{.UnexportedTypeName}}, c.n)

	for _, n := range c.nodes {
		go callGRPC{{.MethodName}}(ctx, n, args, replyChan)
	}

	var (
		replyValues     = make([]*{{.FQRespName}}, 0, c.n)
		reply           = &{{.TypeName}}{NodeIDs: make([]uint32, 0, c.n)}
		clevel      	= LevelNotSet
		rlevel      int
		errCount    int
		quorum      bool
	)

	for {
		select {
		case r := <-replyChan:
			reply.NodeIDs = append(reply.NodeIDs, r.nid)
			if r.err != nil {
				errCount++
				break
			}
			replyValues = append(replyValues, r.reply)
{{- if .QFWithReq}}
			reply.{{.RespName}}, rlevel, quorum = c.qspec.{{.MethodName}}CorrectableQF(args, replyValues)
{{else}}
			reply.{{.RespName}}, rlevel, quorum = c.qspec.{{.MethodName}}CorrectableQF(replyValues)
{{end}}
			if quorum {
				corr.set(reply, rlevel, nil, true)
				return
			}
			if rlevel > clevel {
				clevel = rlevel
				corr.set(reply, rlevel, nil, false)
			}
		case <-ctx.Done():
			corr.set(reply, clevel, QuorumCallError{ctx.Err().Error(), errCount, len(replyValues)}, true)
			return
		}

		if errCount+len(replyValues) == c.n {
			corr.set(reply, clevel, QuorumCallError{"incomplete call", errCount, len(replyValues)}, true)
			return
		}
	}
}

{{- end -}}
{{end}}
`

const mgr_correctable_prelim_tmpl = `
{{/* Remember to run 'make goldenanddev' after editing this file. */}}

{{$pkgName := .PackageName}}

{{if not .IgnoreImports}}
package {{$pkgName}}

import (
	"io"

	"golang.org/x/net/context"
)
{{end}}

{{range $elm := .Services}}

{{if .CorrectablePrelim}}

type {{.UnexportedTypeName}} struct {
	nid   uint32
	reply *{{.FQRespName}}
	err   error
}

func (m *Manager) {{.UnexportedMethodName}}CorrectablePrelim(ctx context.Context, c *Configuration, corr *{{.MethodName}}CorrectablePrelim, args *{{.FQReqName}}) {
	replyChan := make(chan {{.UnexportedTypeName}}, c.n)

	for _, n := range c.nodes {
		go callGRPC{{.MethodName}}Stream(ctx, n, args, replyChan)
	}

	var (
		replyValues = make([]*{{.FQRespName}}, 0, c.n*2)
		reply       = &{{.TypeName}}{NodeIDs: make([]uint32, 0, c.n)}
		clevel      = LevelNotSet
		rlevel      int
		errCount    int
		quorum      bool
	)

	for {
		select {
		case r := <-replyChan:
			reply.NodeIDs = appendIfNotPresent(reply.NodeIDs, r.nid)
			if r.err != nil {
				errCount++
				break
			}
			replyValues = append(replyValues, r.reply)
{{- if .QFWithReq}}
			reply.{{.RespName}}, rlevel, quorum = c.qspec.{{.MethodName}}CorrectablePrelimQF(args, replyValues)
{{else}}
			reply.{{.RespName}}, rlevel, quorum = c.qspec.{{.MethodName}}CorrectablePrelimQF(replyValues)
{{end}}
			if quorum {
				corr.set(reply, rlevel, nil, true)
				return
			}
			if rlevel > clevel {
				clevel = rlevel
				corr.set(reply, rlevel, nil, false)
			}
		case <-ctx.Done():
			corr.set(reply, clevel, QuorumCallError{ctx.Err().Error(), errCount, len(replyValues)}, true)
			return
		}

		if errCount == c.n { // Can't rely on reply count.
			corr.set(reply, clevel, QuorumCallError{"incomplete call", errCount, len(replyValues)}, true)
			return
		}
	}
}

func callGRPC{{.MethodName}}Stream(ctx context.Context, node *Node, args *{{.FQReqName}}, replyChan chan<- {{.UnexportedTypeName}}) {
	x := New{{.ServName}}Client(node.conn)
	y, err := x.{{.MethodName}}(ctx, args)
	if err != nil {
		replyChan <- {{.UnexportedTypeName}}{node.id, nil, err}
		return
	}

	for {
		reply, err := y.Recv()
		if err == io.EOF {
			return
		}
		replyChan <- {{.UnexportedTypeName}}{node.id, reply, err}
		if err != nil {
			return
		}
	}
}

{{- end -}}
{{- end -}}
`

const mgr_multicast_tmpl = `
{{/* Remember to run 'make goldenanddev' after editing this file. */}}

{{$pkgName := .PackageName}}

{{if not .IgnoreImports}}
package {{$pkgName}}

import "golang.org/x/net/context"
{{end}}

{{range $elm := .Services}}

{{if .Multicast}}
func (m *Manager) {{.UnexportedMethodName}}(ctx context.Context, c *Configuration, args *{{.FQReqName}}) error {
	for _, node := range c.nodes {
		go func(n *Node) {
			err := n.{{.MethodName}}Client.Send(args)
			if err == nil {
				return
			}
			if m.logger != nil {
				m.logger.Printf("%d: {{.UnexportedMethodName}} stream send error: %v", n.id, err)
			}
		}(node)
	}

	return nil
}
{{- end -}}
{{end}}
`

const mgr_quorumcall_tmpl = `
{{/* Remember to run 'make goldenanddev' after editing this file. */}}

{{$pkgName := .PackageName}}

{{if not .IgnoreImports}}
package {{$pkgName}}

import (
	"time"

	"golang.org/x/net/context"
	"golang.org/x/net/trace"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)
{{end}}

{{range $elm := .Services}}

{{if or (.QuorumCall) (.Future)}}

type {{.UnexportedTypeName}} struct {
	nid   uint32
	reply *{{.FQRespName}}
	err   error
}

func (m *Manager) {{.UnexportedMethodName}}(ctx context.Context, c *Configuration, {{.MethodArg}}) (r *{{.TypeName}}, err error) {
	var ti traceInfo
	if m.opts.trace {
		ti.tr = trace.New("gorums."+c.tstring()+".Sent", "{{.MethodName}}")
		defer ti.tr.Finish()

		ti.firstLine.cid = c.id
		if deadline, ok := ctx.Deadline(); ok {
			ti.firstLine.deadline = deadline.Sub(time.Now())
		}
		ti.tr.LazyLog(&ti.firstLine, false)

		defer func() {
			ti.tr.LazyLog(&qcresult{
				ids:   r.NodeIDs,
				reply: r.{{.RespName}},
				err:   err,
			}, false)
			if err != nil {
				ti.tr.SetError()
			}
		}()
	}

	replyChan := make(chan {{.UnexportedTypeName}}, c.n)

	if m.opts.trace {
		ti.tr.LazyLog(&payload{sent: true, msg: {{.MethodArgUse}}}, false)
	}

	for _, n := range c.nodes {
		go callGRPC{{.MethodName}}(ctx, n, {{.MethodArgCall}}, replyChan)
	}

	var (
		replyValues = make([]*{{.FQRespName}}, 0, c.n)
		reply       = &{{.TypeName}}{NodeIDs: make([]uint32, 0, c.n)}
		errCount    int
		quorum      bool
	)

	for {
		select {
		case r := <-replyChan:
			reply.NodeIDs = append(reply.NodeIDs, r.nid)
			if r.err != nil {
				errCount++
				break
			}
			if m.opts.trace {
				ti.tr.LazyLog(&payload{sent: false, id: r.nid, msg: r.reply}, false)
			}
			replyValues = append(replyValues, r.reply)
{{- if .QFWithReq}}
			if reply.{{.RespName}}, quorum = c.qspec.{{.MethodName}}QF({{.MethodArgUse}}, replyValues); quorum {
{{else}}
			if reply.{{.RespName}}, quorum = c.qspec.{{.MethodName}}QF(replyValues); quorum {
{{end -}}
				return reply, nil
			}
		case <-ctx.Done():
			return reply, QuorumCallError{ctx.Err().Error(), errCount, len(replyValues)}
		}

		if errCount+len(replyValues) == c.n {
			return reply, QuorumCallError{"incomplete call", errCount, len(replyValues)}
		}
	}
}

func callGRPC{{.MethodName}}(ctx context.Context, node *Node, args *{{.FQReqName}}, replyChan chan<- {{.UnexportedTypeName}}) {
	reply := new({{.FQRespName}})
	start := time.Now()
	err := grpc.Invoke(
		ctx,
		"/{{$pkgName}}.{{.ServName}}/{{.MethodName}}",
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
	replyChan <- {{.UnexportedTypeName}}{node.id, reply, err}
}

{{- end -}}
{{end}}
`

const node_tmpl = `
{{/* Remember to run 'make goldenanddev' after editing this file. */}}

{{- if not .IgnoreImports}}
package {{.PackageName}}

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
)
{{- end}}

// Node encapsulates the state of a node on which a remote procedure call
// can be made.
type Node struct {
	// Only assigned at creation.
	id   uint32
	self bool
	addr string
	conn *grpc.ClientConn


{{range .Clients}}
	{{.}} {{.}}
{{end}}

{{range .Services}}
{{if .Multicast}}
	{{.MethodName}}Client {{.ServName}}_{{.MethodName}}Client
{{end}}
{{end}}

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

{{range .Clients}}
	n.{{.}} = New{{.}}(n.conn)
{{end}}

{{range .Services}}
{{if .Multicast}}
  	n.{{.MethodName}}Client, err = n.{{.ServName}}Client.{{.MethodName}}(context.Background())
  	if err != nil {
  		return fmt.Errorf("stream creation failed: %v", err)
  	}
{{end}}
{{end -}}

	return nil
}

func (n *Node) close() error {
	// TODO: Log error, mainly care about the connection error below.
        // We should log this error, but we currently don't have access to the
        // logger in the manager.
{{- range .Services -}}
{{if .Multicast}}
	_, _ = n.{{.MethodName}}Client.CloseAndRecv()
{{- end -}}
{{end}}
	
	if err := n.conn.Close(); err != nil {
                return fmt.Errorf("conn close error: %v", err)
        }	
	return nil
}
`

const qspec_tmpl = `
{{/* Remember to run 'make goldenanddev' after editing this file. */}}

{{- if not .IgnoreImports}}
package {{.PackageName}}
{{- end}}

// QuorumSpec is the interface that wraps every quorum function.
type QuorumSpec interface {
{{- range $elm := .Services}}
{{- if or (.QuorumCall) (.Future)}}
	// {{.MethodName}}QF is the quorum function for the {{.MethodName}}
	// quorum call method.
{{- if .QFWithReq}}
	{{.MethodName}}QF(req *{{.FQReqName}}, replies []*{{.FQRespName}}) (*{{.FQRespName}}, bool)
{{else}}
	{{.MethodName}}QF(replies []*{{.FQRespName}}) (*{{.FQRespName}}, bool)
{{end}}
{{end}}

{{if .Correctable}}
	// {{.MethodName}}CorrectableQF is the quorum function for the {{.MethodName}}
	// correctable quorum call method.
	{{.MethodName}}CorrectableQF(replies []*{{.FQRespName}}) (*{{.FQRespName}}, int, bool)
{{end}}

{{if .CorrectablePrelim}}
	// {{.MethodName}}CorrectablePrelimQF is the quorum function for the {{.MethodName}} 
	// correctable prelim quourm call method.
	{{.MethodName}}CorrectablePrelimQF(replies []*{{.FQRespName}}) (*{{.FQRespName}}, int, bool)
{{end}}
{{- end -}}
}
`

var templates = map[string]string{
	"config_qc_tmpl":              config_qc_tmpl,
	"mgr_correctable_tmpl":        mgr_correctable_tmpl,
	"mgr_correctable_prelim_tmpl": mgr_correctable_prelim_tmpl,
	"mgr_multicast_tmpl":          mgr_multicast_tmpl,
	"mgr_quorumcall_tmpl":         mgr_quorumcall_tmpl,
	"node_tmpl":                   node_tmpl,
	"qspec_tmpl":                  qspec_tmpl,
}
