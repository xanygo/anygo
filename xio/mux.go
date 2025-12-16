//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-12-08

package xio

import (
	"context"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/xerror"
)

// -------------------- 帧（Frame）格式 --------------------------------
// 头部格式：4 字节 StreamID（大端序） | 1 字节 Flags | 4 字节 Length（大端序）| 4字节 校验码 （大端序）
// 接着是长度为 Length 的 payload。
// --------------------------------------------------------------------

type streamFlag byte

const (
	flagData  streamFlag = 1 << iota // payload 含有正常数据
	flagClose                        // 流被正常关闭；不携带 payload，length == 0
	flagReset                        // 流被重置/异常中止；不携带 payload，length == 0
	flagOpen                         // 显式打开 stream，payload 可携带可不携带，用于提前通知对端
)

const (
	headerSize = 13

	// 将大块写入拆成此大小的 payload，以避免单帧过大
	defaultMaxPayload = 64 * 1024
)

// Mux 在一个 io.ReadWriteCloser 上复用多个 MuxStream。
// 支持并发读写，并通过 MuxStream ID 区分不同逻辑流。
type Mux[T io.ReadWriteCloser] struct {
	parent T

	// 写操作串行化锁，保证多 MuxStream 写入不交叉
	writeMu sync.Mutex

	// 保存所有活跃 MuxStream，通过 MuxStream ID 索引
	streams sync.Map // map[uint32]*MuxStream

	// 下一个分配的 MuxStream ID
	// 作为 client，ID 为偶数 依次为 2,4,6 ...
	// 作为 server，ID 为奇数，依次为 3,5,7 ...
	nextID atomic.Uint32

	// 接收远端发起的 MuxStream
	acceptCh chan *MuxStream[T]

	// Mux 关闭控制
	closeOnce sync.Once

	closedCh chan struct{}

	// 配置项：最大 payload 大小
	maxPayload int
}

// NewMux 创建一个新的 Mux 并启动后台读取循环（readLoop）。
// rw：底层 io.ReadWriteCloser
// chanSize：接收远端 MuxStream 的缓冲通道大小
func NewMux[T io.ReadWriteCloser](client bool, rw T) *Mux[T] {
	m := &Mux[T]{
		parent:     rw,
		acceptCh:   make(chan *MuxStream[T], 32),
		closedCh:   make(chan struct{}),
		maxPayload: defaultMaxPayload,
	}
	if !client {
		m.nextID.Store(1)
	}
	go m.readLoop()
	return m
}

// SetMaxPayload 设置帧的最大 payload 大小，用于分片发送。
func (m *Mux[T]) SetMaxPayload(n int) {
	if n <= 0 {
		return
	}
	m.maxPayload = n
}

func (m *Mux[T]) Unwrap() T {
	return m.parent
}

// Open 创建一个本地发起的 MuxStream。
func (m *Mux[T]) Open() (*MuxStream[T], error) {
	return m.OpenWithPayload(nil)
}

func (m *Mux[T]) OpenWithPayload(payload []byte) (*MuxStream[T], error) {
	select {
	case <-m.closedCh:
		// Mux 已关闭，返回错误
		return nil, errMuxClosed
	default:
	}

	if l := len(payload); l > 0 {
		if l > m.maxPayload {
			return nil, fmt.Errorf("payload too large (%d)", l)
		}
		payload = slices.Clone(payload)
	}
	// 分配新的 MuxStream ID
	id := m.allocID()
	stream := newMuxStream(m, id, false)
	stream.hello = payload

	// 可选：发送 OPEN 帧，让远端提前知道 MuxStream 已创建
	// 不是必须的，因为第一个 DATA 帧也可以隐式创建 MuxStream
	// 但发送 OPEN 更清晰地表明意图
	err := m.writeFrame(id, flagOpen, payload)
	if err == nil {
		m.streams.Store(id, stream)
		return stream, nil
	}
	return nil, err
}

// Accept 阻塞等待远端发起的 MuxStream，或者 Mux 被关闭。
func (m *Mux[T]) Accept() (*MuxStream[T], error) {
	return m.AcceptContext(context.Background())
}

func (m *Mux[T]) AcceptContext(ctx context.Context) (*MuxStream[T], error) {
	select {
	case stream := <-m.acceptCh:
		return stream, nil
	case <-m.closedCh:
		return nil, errMuxClosed
	case <-ctx.Done():
		return nil, context.Cause(ctx)
	}
}

var errMuxClosed = fmt.Errorf("mux %w", xerror.Closed)

// Close 关闭 Mux 及底层 parent 连接。
// 同时会关闭所有活跃的 MuxStream。
func (m *Mux[T]) Close() error {
	return m.doClose(true, io.EOF)
}

func (m *Mux[T]) Range(fn func(s *MuxStream[T]) bool) {
	m.streams.Range(func(_, v any) bool {
		return fn(v.(*MuxStream[T]))
	})
}

func (m *Mux[T]) doClose(notify bool, ce error) error {
	var err error
	m.closeOnce.Do(func() {
		m.closeAllStreams(ce, notify)
		close(m.closedCh)
		err = m.parent.Close()
	})
	m.closeAllStreams(ce, false)
	return err
}

func (m *Mux[T]) closeAllStreams(err error, notify bool) {
	var all []*MuxStream[T]
	m.streams.Range(func(_, v any) bool {
		all = append(all, v.(*MuxStream[T]))
		return true
	})
	for _, stream := range all {
		stream.doClose(flagClose, notify, err)
	}
}

// allocID 分配下一个 MuxStream ID（自动递增）。
// 避免返回 0 作为 MuxStream ID。
func (m *Mux[T]) allocID() uint32 {
	return m.nextID.Add(2)
}

// writeFrame 写入单个帧（包含头部和 payload）。
// 该方法通过加锁保证写操作的串行化，避免多 MuxStream 写入交错。
func (m *Mux[T]) writeFrame(id uint32, flags streamFlag, payload []byte) error {
	select {
	case <-m.closedCh:
		return fmt.Errorf("writeFrame: %w", errMuxClosed)
	default:
	}

	// 构造帧头部
	header := make([]byte, headerSize)
	binary.BigEndian.PutUint32(header[0:4], id)                   // MuxStream ID
	header[4] = byte(flags)                                       // Flags
	binary.BigEndian.PutUint32(header[5:9], uint32(len(payload))) // payload 长度

	crc := crc32.ChecksumIEEE(header[:9])
	binary.BigEndian.PutUint32(header[9:13], crc) // CRC32

	m.writeMu.Lock()
	defer m.writeMu.Unlock()

	var err error
	_, err = m.parent.Write(header)     // 写入头部
	if err == nil && len(payload) > 0 { // 写入 payload（如果有）
		_, err = m.parent.Write(payload)
	}

	// 写通道异常，关闭整个连接
	if err != nil {
		go m.doClose(false, err)
	}
	return err
}

// writeData 对大块数据进行分片写入。
// 将超过 maxPayload 的数据拆分为多个帧发送。
func (m *Mux[T]) writeData(id uint32, p []byte) error {
	num := m.maxPayload
	for len(p) > 0 {
		chunk := p
		if len(chunk) > num {
			chunk = chunk[:num]
		}
		if err := m.writeFrame(id, flagData, chunk); err != nil {
			return err
		}
		p = p[len(chunk):]
	}
	return nil
}

// readLoop 后台循环读取帧，并分发到对应的 MuxStream。
func (m *Mux[T]) readLoop() {
	header := make([]byte, headerSize)
	for {
		// 读取帧头
		if _, err := io.ReadFull(m.parent, header); err != nil {
			m.doClose(false, fmt.Errorf("read header %w", err))
			return
		}

		// 解析头部字段
		id := binary.BigEndian.Uint32(header[0:4])     // MuxStream ID
		flags := streamFlag(header[4])                 // Flags
		length := binary.BigEndian.Uint32(header[5:9]) // payload 长度

		crcRecv := binary.BigEndian.Uint32(header[9:13])
		crcExpect := crc32.ChecksumIEEE(header[:9])

		if crcExpect != crcRecv {
			ce := fmt.Errorf("invalid stream sum, sid=%d", id)
			m.doClose(false, ce)
			return
		}

		var payload []byte
		if length > 0 {
			payload = make([]byte, length)
			if _, err := io.ReadFull(m.parent, payload); err != nil {
				m.doClose(false, fmt.Errorf("read payload %w", err))
				return
			}
		}

		var stream *MuxStream[T]
		v, ok := m.streams.Load(id)
		if ok {
			if flags == flagOpen {
				err := fmt.Errorf("streamID=%d already exists", id)
				m.doClose(false, err)
				return
			}
			stream = v.(*MuxStream[T])
		} else if flags == flagOpen {
			// 新的远端发起 MuxStream
			stream = newMuxStream[T](m, id, true)
			stream.hello = payload

			// 存储后再推送到 acceptCh
			m.streams.Store(id, stream)
			// 阻塞推送到 acceptCh,必须被接受到
			select {
			case m.acceptCh <- stream:
			case <-m.closedCh:
				return
			}
		} else {
			err := fmt.Errorf("invalid streaID(%d) or flag(%b)", id, flags)
			m.doClose(false, err)
			return
		}

		switch flags {
		case flagOpen:
			// do nothing
		case flagData:
			// DATA：将 payload 放入 MuxStream 的读取缓冲区
			stream.push(payload)
		case flagClose:
			// CLOSE：远端正常关闭 MuxStream
			stream.doClose(flagClose, false, io.EOF)
		case flagReset:
			// RESET：远端复位/中止 MuxStream
			stream.doClose(flagReset, false, io.EOF)
		default:
			ce := fmt.Errorf("unknown flags %d", flags)
			m.doClose(false, ce)
			return
		}
	}
}

// MuxStream 表示一个复用通道（逻辑流），在 Mux 上进行全双工读写。
// 支持本地或远端发起，读写操作可并发。
type MuxStream[T io.ReadWriteCloser] struct {
	id       uint32    // MuxStream ID
	mux      *Mux[T]   // 所属的 Mux
	remote   bool      // 是否由远端发起
	createAt time.Time // 创建时间
	hello    []byte    // Open 创建时，携带的数据

	// 读相关
	mu sync.Mutex // 保护 pending 的状态

	pending []byte // 待消费的未读数据

	ch chan []byte // 传递数据块的通道

	// 关闭状态
	closed atomic.Bool // 本地或远端是否关闭

	closedCh  chan struct{}      // 用于通知关闭
	closeOnce sync.Once          // 确保关闭操作只执行一次
	closeErr  xsync.Value[error] // 关闭时候的错误
}

// newMuxStream 创建一个新的 MuxStream。
// m：所属 Mux
// id：MuxStream ID
// remote：是否远端发起
func newMuxStream[T io.ReadWriteCloser](m *Mux[T], id uint32, remote bool) *MuxStream[T] {
	return &MuxStream[T]{
		id:       id,
		mux:      m,
		remote:   remote,
		ch:       make(chan []byte, 64), // 默认缓冲区大小
		closedCh: make(chan struct{}),   // 初始化关闭通知通道
		createAt: time.Now(),
	}
}

func (s *MuxStream[T]) Parent() *Mux[T] {
	return s.mux
}

// Hello 创建 stream 时候，携带的数据
func (s *MuxStream[T]) Hello() []byte {
	return s.hello
}

// ID 返回 MuxStream 的唯一标识
func (s *MuxStream[T]) ID() uint32 {
	return s.id
}

// CreateAt 创建时间
func (s *MuxStream[T]) CreateAt() time.Time {
	return s.createAt
}

// Read 实现 io.Reader 接口。
// 当 MuxStream 已关闭且没有剩余数据时返回 io.EOF。
func (s *MuxStream[T]) Read(p []byte) (int, error) {
	// 快路径：如果 pending 缓冲有数据，直接读取
	s.mu.Lock()
	if len(s.pending) > 0 {
		n := copy(p, s.pending)
		s.pending = s.pending[n:]
		// 如果 pending 已空，置为 nil 以释放内存
		if len(s.pending) == 0 {
			s.pending = nil
		}
		s.mu.Unlock()
		return n, nil
	}
	s.mu.Unlock()

	// 如果已关闭且没有数据
	if s.closed.Load() {
		return 0, io.EOF
	}

	// 阻塞等待新的数据块或 MuxStream 被关闭
	select {
	case data, ok := <-s.ch:
		if !ok {
			return 0, io.EOF
		}
		// 尽可能复制请求长度的数据，如果 data 更长，将剩余部分放入 pending
		n := copy(p, data)
		if n < len(data) {
			s.mu.Lock()
			s.pending = append(s.pending, data[n:]...)
			s.mu.Unlock()
		}
		return n, nil
	case <-s.closedCh:
		return 0, io.EOF
	case <-s.mux.closedCh:
		return 0, io.EOF
		// case <-time.After(10 * time.Second):
		//	return 0, fmt.Errorf("timeout")
	}
}

// Write 实现 io.Writer 接口。
// 如果数据过大，会自动分片通过 Mux 发送。
func (s *MuxStream[T]) Write(p []byte) (int, error) {
	if s.closed.Load() {
		return 0, io.EOF
	}

	// 通过 Mux 写入分片数据
	if err := s.mux.writeData(s.id, p); err != nil {
		return 0, err
	}
	return len(p), nil
}

func (s *MuxStream[T]) Done() <-chan struct{} {
	return s.closedCh
}

func (s *MuxStream[T]) Err() error {
	return s.closeErr.Load()
}

// push 由 Mux.readLoop 调用，用于向 MuxStream 传入数据。
func (s *MuxStream[T]) push(data []byte) {
	if s.closed.Load() || len(data) == 0 {
		return
	}
	select {
	case s.ch <- data:
		// 缓冲区满时，阻塞
	case <-s.closedCh:
	case <-s.mux.closedCh:
	}
}

func (s *MuxStream[T]) doClose(flag streamFlag, notify bool, err error) error {
	isCloseFlag := flag&flagClose != 0 || flag&flagReset != 0
	if !isCloseFlag {
		panic(fmt.Errorf("invalid flag %v", flag))
	}

	s.closeOnce.Do(func() {
		if notify { // 是否需要给远端发送已经关闭的消息
			_ = s.mux.writeFrame(s.id, flag, nil) // 尝试发送 Close 消息
		}

		if err == nil {
			err = io.EOF
		}
		// 标记本地已关闭，并从 Mux 中移除
		s.closeErr.Store(err)
		s.closed.Store(true)

		close(s.closedCh)
		close(s.ch)

		s.mux.streams.Delete(s.id)
	})
	return nil
}

// Close 本地关闭 MuxStream，并发送 CLOSE 帧通知远端。
func (s *MuxStream[T]) Close() error {
	return s.doClose(flagClose, true, io.EOF)
}
