package messaging

import (
	"errors"
	"net/http"
	"ride-sharing/shared/contracts"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	ErrConnectionNotFound = errors.New("connection not found")
)

type connWrapper struct {
	connection *websocket.Conn
	mutex      sync.Mutex
}
type ConnectionManager struct {
	mutex       sync.RWMutex
	connections map[string]*connWrapper
}

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]*connWrapper),
	}
}

func (c *ConnectionManager) Upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (c *ConnectionManager) Add(conn *websocket.Conn, userId string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.connections[userId] = &connWrapper{
		connection: conn,
		mutex:      sync.Mutex{},
	}

}

func (c *ConnectionManager) Remove(id string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.connections, id)
}

func (c *ConnectionManager) Get(id string) (*websocket.Conn, bool) {
	c.mutex.RLock()

	conn, exists := c.connections[id]

	if !exists {
		return nil, false
	}

	return conn.connection, true

}

func (c *ConnectionManager) SendMessage(id string, message contracts.WSMessage) error {
	c.mutex.Lock()

	defer c.mutex.Unlock()

	conn, exists := c.connections[id]

	if !exists {
		return ErrConnectionNotFound
	}

	conn.mutex.Lock()
	defer conn.mutex.Unlock()

	return conn.connection.WriteJSON(message)

}
