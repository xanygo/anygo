//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-30

package xpool

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

func New[V io.Closer](opt *Option, ct Factory[V]) Pool[V] {
	if opt == nil {
		opt = &Option{}
	}
	ctx, cancel := context.WithCancel(context.Background())
	pool := &simple[V]{
		maxOpen:     opt.MaxOpen,
		maxLifetime: opt.MaxLifeTime,
		maxIdleTime: opt.MaxIdleTime,
		creator:     ct,
		stop:        cancel,
	}
	go pool.connectionOpener(ctx)
	return pool
}

var _ finalCloser = (*element[net.Conn])(nil)

var _ Entry[net.Conn] = (*element[net.Conn])(nil)

type element[V io.Closer] struct {
	id        uint64
	pool      *simple[V]
	createdAt time.Time
	validator Validator[V]

	mux         sync.Mutex // guards following
	obj         V
	closed      bool
	finalClosed bool      // obj.Close has been called
	lastUsedAt  time.Time // 上次使用时间
	usageCount  uint64    // 已被使用次数，回池后+1

	// guarded by pool.mu
	inUse bool

	returnedAt time.Time // 返回对象池的时间或者创建时间
	onPut      []func()  // code (with pool.mu held) run when conn is next returned
}

func (dc *element[V]) ID() uint64 {
	return dc.id
}

func (dc *element[V]) Object() V {
	return dc.obj
}

func (dc *element[V]) CreatedAt() time.Time {
	return dc.createdAt
}

func (dc *element[V]) LastUsedAt() time.Time {
	dc.mux.Lock()
	defer dc.mux.Unlock()
	return dc.lastUsedAt
}

func (dc *element[V]) UsageCount() uint64 {
	dc.mux.Lock()
	defer dc.mux.Unlock()
	return dc.usageCount
}

func (dc *element[V]) Release(err error) {
	dc.releaseConn(err)
}

func (dc *element[V]) withLock(fn func()) {
	dc.mux.Lock()
	defer dc.mux.Unlock()
	fn()
}

// finalClose 在 放回 Pool 之后，需要关闭原始对象时调用
func (dc *element[V]) finalClose() error {
	var err error
	dc.withLock(func() {
		dc.finalClosed = true
		err = dc.obj.Close()
		var emp V
		dc.obj = emp
	})
	dc.pool.withLock(func() {
		dc.pool.numOpen--
		dc.pool.maybeOpenNewConnections()
	})
	dc.pool.numClosed.Add(1)
	return err
}

func (dc *element[V]) releaseConn(err error) {
	dc.pool.putConn(dc, err)
}

func (dc *element[V]) expired(timeout time.Duration) bool {
	if timeout <= 0 {
		return false
	}
	return dc.createdAt.Add(timeout).Before(time.Now())
}

// validateConnection 验证是否有效
func (dc *element[V]) validateConnection() bool {
	if dc.validator == nil {
		return true
	}
	return dc.validator.Validate(dc.obj) == nil
}

// the dc.pool's Mutex is held.
func (dc *element[V]) closeDBLocked() func() error {
	dc.mux.Lock()
	defer dc.mux.Unlock()
	if dc.closed {
		return func() error { return errors.New("duplicate element close") }
	}
	dc.closed = true
	return dc.pool.removeDepLocked(dc, dc)
}

func (dc *element[V]) Close() error {
	dc.mux.Lock()
	if dc.closed {
		dc.mux.Unlock()
		return nil
	}
	dc.closed = true
	dc.mux.Unlock() // not defer; removeDep finalClose calls may need to lock

	// And now updates that require holding dc.mu.Lock.
	dc.pool.mu.Lock()
	fn := dc.pool.removeDepLocked(dc, dc)
	dc.pool.mu.Unlock()
	return fn()
}

var globalID atomic.Uint64

var _ Pool[net.Conn] = (*simple[net.Conn])(nil)

type simple[V io.Closer] struct {
	// Total time waited for new connections.
	waitDuration atomic.Int64

	// numClosed is an atomic counter which represents a total number of
	// closed connections. Stmt.openStmt checks it before cleaning closed
	// connections in Stmt.css.
	numClosed atomic.Uint64

	creator Factory[V]

	mu           sync.Mutex        // protects following fields
	frees        []*element[V]     // free connections ordered by returnedAt oldest to newest
	connRequests connRequestSet[V] // 排队请求
	numOpen      int               // number of opened and pending open connections

	// Used to signal the need for new connections
	// a goroutine running connectionOpener() reads on this chan and
	// maybeOpenNewConnections sends on the chan (one send per needed connection)
	// It is closed during pool.Close(). The close tells the connectionOpener
	// goroutine to exit.
	openerCh chan struct{}

	closed            bool
	dep               map[finalCloser]depSet
	lastPut           map[*element[V]]string // stacktrace of last conn's put; debug only
	maxIdleCount      int                    // zero means defaultMaxIdleConns; negative means 0
	maxOpen           int                    // <= 0 means unlimited
	maxLifetime       time.Duration          // maximum amount of time a connection may be reused
	maxIdleTime       time.Duration          // maximum amount of time a connection may be idle before being closed
	cleanerCh         chan struct{}
	waitCount         int64 // Total number of connections waited for.
	maxIdleClosed     int64 // Total number of connections closed due to idle count.
	maxIdleTimeClosed int64 // Total number of connections closed due to idle time.
	maxLifetimeClosed int64 // Total number of connections closed due to max connection lifetime limit.

	stop func()
}

func (p *simple[V]) Get(ctx context.Context) (Entry[V], error) {
	var err error
	var el *element[V]
	err = p.retry(func(strategy connReuseStrategy) error {
		el, err = p.conn(ctx, strategy)
		return err
	})
	return el, err
}

func (p *simple[V]) Stats() Stats {
	p.mu.Lock()
	defer p.mu.Unlock()
	st := Stats{
		Open:              !p.closed,
		NumOpen:           p.numOpen,
		InUse:             p.numOpen - len(p.frees),
		Idle:              len(p.frees),
		WaitCount:         p.waitCount,
		WaitDuration:      time.Duration(p.waitDuration.Load()),
		MaxIdleClosed:     p.maxIdleClosed,
		MaxIdleTimeClosed: p.maxIdleTimeClosed,
		MaxLifeTimeClosed: p.maxLifetimeClosed,
	}
	return st
}

func (p *simple[V]) withLock(fn func()) {
	p.mu.Lock()
	defer p.mu.Unlock()
	fn()
}

// connReuseStrategy determines how (*DB).conn returns database connections.
type connReuseStrategy uint8

const (
	// alwaysNewConn forces a new connection to the database.
	alwaysNewConn connReuseStrategy = iota

	// cachedOrNewConn returns a cached connection, if available, else waits
	// for one to become available (if MaxOpenConns has been reached) or
	// creates a new database connection.
	cachedOrNewConn
)

func (p *simple[V]) newElement(obj V) *element[V] {
	now := time.Now()
	vat, _ := p.creator.(Validator[V])
	return &element[V]{
		id:         globalID.Add(1),
		pool:       p,
		createdAt:  now,
		returnedAt: now,
		obj:        obj,
		validator:  vat,
	}
}

// maxBadEntryRetries is the number of maximum retries if the driver returns
// driver.ErrBadConn to signal a broken connection before forcing a new
// connection to be opened.
const maxBadEntryRetries = 2

func (p *simple[V]) retry(fn func(strategy connReuseStrategy) error) error {
	for i := int64(0); i < maxBadEntryRetries; i++ {
		err := fn(cachedOrNewConn)
		// retry if err is ErrBadEntry
		if err == nil || !errors.Is(err, ErrBadEntry) {
			return err
		}
	}

	return fn(alwaysNewConn)
}

// conn returns a newly-opened or cached *driverConn.
func (p *simple[V]) conn(ctx context.Context, strategy connReuseStrategy) (*element[V], error) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil, ErrClosed
	}
	// Check if the context is expired.
	select {
	default:
	case <-ctx.Done():
		p.mu.Unlock()
		return nil, ctx.Err()
	}
	lifetime := p.maxLifetime

	// Prefer a free connection, if possible.
	last := len(p.frees) - 1
	if strategy == cachedOrNewConn && last >= 0 {
		// Reuse the lowest idle time connection so we can close
		// connections which remain idle as soon as possible.
		conn := p.frees[last]
		p.frees = p.frees[:last]
		conn.inUse = true
		if conn.expired(lifetime) {
			p.maxLifetimeClosed++
			p.mu.Unlock()
			conn.Close()
			return nil, ErrBadEntry
		}
		p.mu.Unlock()

		return conn, nil
	}

	// Out of free connections or we were asked not to use one. If we're not
	// allowed to open any more connections, make a request and wait.
	if p.maxOpen > 0 && p.numOpen >= p.maxOpen {
		// Make the connRequest channel. It's buffered so that the
		// connectionOpener doesn't block while waiting for the req to be read.
		req := make(chan connRequest[V], 1)
		delHandle := p.connRequests.Add(req)
		p.waitCount++
		p.mu.Unlock()

		waitStart := time.Now()

		// Timeout the connection request with the context.
		select {
		case <-ctx.Done():
			// Remove the connection request and ensure no value has been sent
			// on it after removing.
			p.mu.Lock()
			deleted := p.connRequests.Delete(delHandle)
			p.mu.Unlock()

			p.waitDuration.Add(int64(time.Since(waitStart)))

			// If we failed to delete it, that means either the DB was closed or
			// something else grabbed it and is about to send on it.
			if !deleted {
				select {
				default:
				case ret, ok := <-req:
					if ok && ret.conn != nil {
						p.putConn(ret.conn, ret.err)
					}
				}
			}
			return nil, ctx.Err()
		case ret, ok := <-req:
			p.waitDuration.Add(int64(time.Since(waitStart)))

			if !ok {
				return nil, ErrClosed
			}
			if strategy == cachedOrNewConn && ret.err == nil && ret.conn.expired(lifetime) {
				p.mu.Lock()
				p.maxLifetimeClosed++
				p.mu.Unlock()
				ret.conn.Close()
				return nil, ErrBadEntry
			}
			if ret.conn == nil {
				return nil, ret.err
			}

			return ret.conn, ret.err
		}
	}

	p.numOpen++ // optimistically
	p.mu.Unlock()
	ci, err := p.creator.New(ctx)
	if err != nil {
		p.mu.Lock()
		p.numOpen-- // correct for earlier optimism
		p.maybeOpenNewConnections()
		p.mu.Unlock()
		return nil, err
	}
	p.mu.Lock()
	dc := p.newElement(ci)
	dc.inUse = true

	p.addDepLocked(dc, dc)
	p.mu.Unlock()
	return dc, nil
}

// debugGetPut determines whether getConn & putConn calls' stack traces
// are returned for more verbose crashes.
const debugGetPut = false

func stack() string {
	var buf [2 << 10]byte
	return string(buf[:runtime.Stack(buf[:], false)])
}

// putConn adds a connection to the pool's free simple.
// err is optionally the last error that occurred on this connection.
func (p *simple[V]) putConn(dc *element[V], err error) {
	if !errors.Is(err, ErrBadEntry) {
		if !dc.validateConnection() {
			err = ErrBadEntry
		}
	}
	p.mu.Lock()
	if !dc.inUse {
		p.mu.Unlock()
		return
	}

	if !errors.Is(err, ErrBadEntry) && dc.expired(p.maxLifetime) {
		p.maxLifetimeClosed++
		err = ErrBadEntry
	}
	if debugGetPut {
		p.lastPut[dc] = stack()
	}
	dc.inUse = false
	dc.returnedAt = time.Now()

	for _, fn := range dc.onPut {
		fn()
	}
	dc.onPut = nil

	if errors.Is(err, ErrBadEntry) {
		// Don't reuse bad connections.
		// Since the conn is considered bad and is being discarded, treat it
		// as closed. Don't decrement the open count here, finalClose will
		// take care of that.
		p.maybeOpenNewConnections()
		p.mu.Unlock()
		dc.Close()
		return
	}
	added := p.putConnDBLocked(dc, nil)
	p.mu.Unlock()

	if !added {
		dc.Close()
		return
	}
}

// Satisfy a connRequest or put the driverConn in the idle simple and return true
// or return false.
// putConnDBLocked will satisfy a connRequest if there is one, or it will
// return the *driverConn to the freeConn list if err == nil and the idle
// connection limit will not be exceeded.
// If err != nil, the value of dc is ignored.
// If err == nil, then dc must not equal nil.
// If a connRequest was fulfilled or the *driverConn was placed in the
// freeConn list, then true is returned, otherwise false is returned.
func (p *simple[V]) putConnDBLocked(dc *element[V], err error) bool {
	if p.closed {
		return false
	}
	if p.maxOpen > 0 && p.numOpen > p.maxOpen {
		return false
	}
	if req, ok := p.connRequests.TakeRandom(); ok {
		if err == nil {
			dc.inUse = true
		}
		req <- connRequest[V]{
			conn: dc,
			err:  err,
		}
		return true
	} else if err == nil && !p.closed {
		if p.maxIdleConnsLocked() > len(p.frees) {
			p.frees = append(p.frees, dc)
			p.startCleanerLocked()
			return true
		}
		p.maxIdleClosed++
	}
	return false
}

const defaultMaxIdleConns = 2

func (p *simple[V]) maxIdleConnsLocked() int {
	n := p.maxIdleCount
	switch {
	case n == 0:
		return defaultMaxIdleConns
	case n < 0:
		return 0
	default:
		return n
	}
}

// startCleanerLocked starts connectionCleaner if needed.
func (p *simple[V]) startCleanerLocked() {
	if (p.maxLifetime > 0 || p.maxIdleTime > 0) && p.numOpen > 0 && p.cleanerCh == nil {
		p.cleanerCh = make(chan struct{}, 1)
		go p.connectionCleaner(p.shortestIdleTimeLocked())
	}
}

func (p *simple[V]) shortestIdleTimeLocked() time.Duration {
	if p.maxIdleTime <= 0 {
		return p.maxLifetime
	}
	if p.maxLifetime <= 0 {
		return p.maxIdleTime
	}
	return min(p.maxIdleTime, p.maxLifetime)
}

func (p *simple[V]) connectionCleaner(d time.Duration) {
	const minInterval = time.Second

	if d < minInterval {
		d = minInterval
	}
	t := time.NewTimer(d)

	for {
		select {
		case <-t.C:
		case <-p.cleanerCh: // maxLifetime was changed or pool was closed.
		}

		p.mu.Lock()

		d = p.shortestIdleTimeLocked()
		if p.closed || p.numOpen == 0 || d <= 0 {
			p.cleanerCh = nil
			p.mu.Unlock()
			return
		}

		d, closing := p.connectionCleanerRunLocked(d)
		p.mu.Unlock()
		for _, c := range closing {
			c.Close()
		}

		if d < minInterval {
			d = minInterval
		}

		if !t.Stop() {
			select {
			case <-t.C:
			default:
			}
		}
		t.Reset(d)
	}
}

// connectionCleanerRunLocked removes connections that should be closed from
// freeConn and returns them along side an updated duration to the next check
// if a quicker check is required to ensure connections are checked appropriately.
func (p *simple[V]) connectionCleanerRunLocked(d time.Duration) (time.Duration, []*element[V]) {
	var idleClosing int64
	var closing []*element[V]
	if p.maxIdleTime > 0 {
		// As freeConn is ordered by returnedAt process
		// in reverse order to minimise the work needed.
		idleSince := time.Now().Add(-p.maxIdleTime)
		last := len(p.frees) - 1
		for i := last; i >= 0; i-- {
			c := p.frees[i]
			if c.returnedAt.Before(idleSince) {
				i++
				closing = p.frees[:i:i]
				p.frees = p.frees[i:]
				idleClosing = int64(len(closing))
				p.maxIdleTimeClosed += idleClosing
				break
			}
		}

		if len(p.frees) > 0 {
			c := p.frees[0]
			if d2 := c.returnedAt.Sub(idleSince); d2 < d {
				// Ensure idle connections are cleaned up as soon as
				// possible.
				d = d2
			}
		}
	}

	if p.maxLifetime > 0 {
		expiredSince := time.Now().Add(-p.maxLifetime)
		for i := 0; i < len(p.frees); i++ {
			c := p.frees[i]
			if c.createdAt.Before(expiredSince) {
				closing = append(closing, c)

				last := len(p.frees) - 1
				// Use slow delete as order is required to ensure
				// connections are reused least idle time first.
				copy(p.frees[i:], p.frees[i+1:])
				p.frees[last] = nil
				p.frees = p.frees[:last]
				i--
			} else if d2 := c.createdAt.Sub(expiredSince); d2 < d {
				// Prevent connections sitting the freeConn when they
				// have expired by updating our next deadline d.
				d = d2
			}
		}
		p.maxLifetimeClosed += int64(len(closing)) - idleClosing
	}

	return d, closing
}

// Runs in a separate goroutine, opens new connections when requested.
func (p *simple[V]) connectionOpener(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-p.openerCh:
			p.openNewConnection(ctx)
		}
	}
}

// Open one new connection
func (p *simple[V]) openNewConnection(ctx context.Context) {
	// maybeOpenNewConnections has already executed pool.numOpen++ before it sent
	// on pool.openerCh. This function must execute pool.numOpen-- if the
	// connection fails or is closed before returning.
	ci, err := p.creator.New(ctx)
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		if err == nil {
			ci.Close()
		}
		p.numOpen--
		return
	}
	if err != nil {
		p.numOpen--
		p.putConnDBLocked(nil, err)
		p.maybeOpenNewConnections()
		return
	}
	dc := p.newElement(ci)
	if p.putConnDBLocked(dc, err) {
		p.addDepLocked(dc, dc)
	} else {
		p.numOpen--
		ci.Close()
	}
}

// Assumes pool.mu is locked.
// If there are connRequests and the connection limit hasn't been reached,
// then tell the connectionOpener to open new connections.
func (p *simple[V]) maybeOpenNewConnections() {
	numRequests := p.connRequests.Len()
	if p.maxOpen > 0 {
		numCanOpen := p.maxOpen - p.numOpen
		if numRequests > numCanOpen {
			numRequests = numCanOpen
		}
	}
	for numRequests > 0 {
		p.numOpen++ // optimistically
		numRequests--
		if p.closed {
			return
		}
		p.openerCh <- struct{}{}
	}
}

// addDep notes that x now depends on dep, and x's finalClose won't be
// called until all of x's dependencies are removed with removeDep.
//
//	func (p *simple[V]) addDep(x finalCloser, dep any) {
//		p.mu.Lock()
//		defer p.mu.Unlock()
//		p.addDepLocked(x, dep)
//	}
func (p *simple[V]) addDepLocked(x finalCloser, dep any) {
	if p.dep == nil {
		p.dep = make(map[finalCloser]depSet)
	}
	xdep := p.dep[x]
	if xdep == nil {
		xdep = make(depSet)
		p.dep[x] = xdep
	}
	xdep[dep] = true
}

func (p *simple[V]) Put(e Entry[V], err error) {
	p.putConn(e.(*element[V]), err)
}

func (p *simple[V]) removeDepLocked(x finalCloser, dep any) func() error {
	xdep, ok := p.dep[x]
	if !ok {
		panic(fmt.Sprintf("unpaired removeDep: no deps for %T", x))
	}

	l0 := len(xdep)
	delete(xdep, dep)

	switch len(xdep) {
	case l0:
		// Nothing removed. Shouldn't happen.
		panic(fmt.Sprintf("unpaired removeDep: no %T dep on %T", dep, x))
	case 0:
		// No more dependencies.
		delete(p.dep, x)
		return x.finalClose
	default:
		// Dependencies remain.
		return func() error { return nil }
	}
}

func (p *simple[V]) Close() error {
	p.mu.Lock()
	if p.closed { // Make p.Close idempotent
		p.mu.Unlock()
		return nil
	}
	if p.cleanerCh != nil {
		close(p.cleanerCh)
	}
	var err error
	fns := make([]func() error, 0, len(p.frees))
	for _, dc := range p.frees {
		fns = append(fns, dc.closeDBLocked())
	}
	p.frees = nil
	p.closed = true
	p.connRequests.CloseAndRemoveAll()
	p.mu.Unlock()

	for _, fn := range fns {
		err1 := fn()
		if err1 != nil {
			err = err1
		}
	}

	p.stop()
	return err
}

// -------------------------------------------
// connRequest represents one request for a new connection
// When there are no idle connections available, DB.conn will create
// a new connRequest and put it on the pool.connRequests list.
type connRequest[V io.Closer] struct {
	conn *element[V]
	err  error
}

// depSet is a finalCloser's outstanding dependencies
type depSet map[any]bool // set of true bools

// The finalCloser interface is used by (*DB).addDep and related
// dependency reference counting.
type finalCloser interface {
	// finalClose is called when the reference count of an object
	// goes to zero. (*DB).mu is not held while calling it.
	finalClose() error
}

type connRequestSet[V io.Closer] struct {
	// s are the elements in the set.
	s []connRequestAndIndex[V]
}

// Add adds v to the set of waiting requests.
// The returned connRequestDelHandle can be used to remove the item from
// the set.
func (s *connRequestSet[V]) Add(v chan connRequest[V]) connRequestDelHandle {
	idx := len(s.s)
	// TODO(bradfitz): for simplicity, this always allocates a new int-sized
	// allocation to store the index. But generally the set will be small and
	// under a scannable-threshold. As an optimization, we could permit the *int
	// to be nil when the set is small and should be scanned. This works even if
	// the set grows over the threshold with delete handles outstanding because
	// an element can only move to a lower index. So if it starts with a nil
	// position, it'll always be in a low index and thus scannable. But that
	// can be done in a follow-up change.
	idxPtr := &idx
	s.s = append(s.s, connRequestAndIndex[V]{req: v, curIdx: idxPtr})
	return connRequestDelHandle{idx: idxPtr}
}

// connRequestDelHandle is an opaque handle to delete an
// item from calling Add.
type connRequestDelHandle struct {
	idx *int // pointer to index; or -1 if not in slice
}

// Len returns the length of the set.
func (s *connRequestSet[V]) Len() int { return len(s.s) }

// Delete removes an element from the set.
//
// It reports whether the element was deleted. (It can return false if a caller
// of TakeRandom took it meanwhile, or upon the second call to Delete)
func (s *connRequestSet[V]) Delete(h connRequestDelHandle) bool {
	idx := *h.idx
	if idx < 0 {
		return false
	}
	s.deleteIndex(idx)
	return true
}

func (s *connRequestSet[V]) deleteIndex(idx int) {
	// Mark item as deleted.
	*(s.s[idx].curIdx) = -1
	// Copy last element, updating its position
	// to its new home.
	if idx < len(s.s)-1 {
		last := s.s[len(s.s)-1]
		*last.curIdx = idx
		s.s[idx] = last
	}
	// Zero out last element (for GC) before shrinking the slice.
	s.s[len(s.s)-1] = connRequestAndIndex[V]{}
	s.s = s.s[:len(s.s)-1]
}

// CloseAndRemoveAll closes all channels in the set
// and clears the set.
func (s *connRequestSet[V]) CloseAndRemoveAll() {
	for _, v := range s.s {
		*v.curIdx = -1
		close(v.req)
	}
	s.s = nil
}

// TakeRandom returns and removes a random element from s
// and reports whether there was one to take. (It returns ok=false
// if the set is empty.)
func (s *connRequestSet[V]) TakeRandom() (v chan connRequest[V], ok bool) {
	if len(s.s) == 0 {
		return nil, false
	}
	pick := rand.IntN(len(s.s))
	e := s.s[pick]
	s.deleteIndex(pick)
	return e.req, true
}

type connRequestAndIndex[V io.Closer] struct {
	// req is the element in the set.
	req chan connRequest[V]

	// curIdx points to the current location of this element in
	// connRequestSet.s. It gets set to -1 upon removal.
	curIdx *int
}
