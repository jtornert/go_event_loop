package runtime

import (
	"context"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/starfederation/datastar-go/datastar"
)

const TicksPerSecond = 10

type EventLoopId struct{}

type ConnectionHandle struct {
	idx uint8
	gen uint8
}

type Connection struct {
	Sse            *datastar.ServerSentEventGenerator
	Request        *http.Request
	LastUpdateTick uint64
	Handle         ConnectionHandle
	Busy           bool
}

type ConnectionMap struct {
	Connections []Connection
	Unused      []bool
	Length      uint8
	MaxLength   uint8
	Mutex       sync.RWMutex
}

func NewConnectionMap(maxLength uint8) ConnectionMap {
	unused := make([]bool, maxLength)
	unused[0] = false
	for i := uint8(1); i < maxLength; i += 1 {
		unused[i] = true
	}
	return ConnectionMap{
		Connections: make([]Connection, maxLength),
		Unused:      unused,
		MaxLength:   maxLength,
		Length:      1,
	}
}

func (cm *ConnectionMap) Add(conn Connection) (ConnectionHandle, bool) {
	cm.Mutex.Lock()
	defer cm.Mutex.Unlock()
	if cm.Length == cm.MaxLength {
		return ConnectionHandle{}, false
	}
	for i := uint8(1); i < cm.MaxLength; i += 1 {
		if cm.Unused[i] {
			cm.Unused[i] = false
			prevGen := cm.Connections[i].Handle.gen
			conn.Handle.idx = uint8(i)
			conn.Handle.gen = prevGen + 1
			cm.Connections[i] = conn
			cm.Length += 1
			return conn.Handle, true
		}
	}
	return ConnectionHandle{}, false
}

func (cm *ConnectionMap) Get(handle ConnectionHandle) *Connection {
	cm.Mutex.RLock()
	defer cm.Mutex.RUnlock()
	idx := handle.idx
	conn := cm.Connections[idx]
	if !cm.Unused[idx] && conn.Handle.gen == handle.gen {
		return &cm.Connections[idx]
	}
	return nil
}

func (cm *ConnectionMap) Remove(handle ConnectionHandle) bool {
	cm.Mutex.Lock()
	defer cm.Mutex.Unlock()
	idx := handle.idx
	if !cm.Unused[idx] && cm.Connections[idx].Handle.gen == handle.gen {
		cm.Unused[idx] = true
		cm.Length -= 1
		return true
	}
	return false
}

type ConnectionCallback func(uint64)

var eventHandlers map[*EventLoopId]ConnectionCallback

func AddEventHandler(loopId *EventLoopId, cb ConnectionCallback) {
	eventHandlers[loopId] = cb
}

func Init() {
	eventHandlers = make(map[*EventLoopId]ConnectionCallback)
}

func RunEventLoop(ctx context.Context) {
	for i := range uint64(math.MaxUint64) {
		now := time.Now()
		select {
		case <-ctx.Done():
			return
		default:
			for _, handler := range eventHandlers {
				handler(i)
			}
			time.Sleep(time.Second/TicksPerSecond - time.Since(now))
		}
	}
}
