package lobbyservice

import (
	"fmt"
	"net/http"
	"sync"
)

type Connection struct {
	// Данные о соединении
	Name    string
	Writter *http.ResponseWriter
}

type ConnectionStack struct {
	connections []Connection
	mutex       sync.Mutex
}

func (cs *ConnectionStack) ObserveLobby(ls *Lobby) {
	go func() {
		for {
			conn, ok := cs.Pop()
			if ok {
				fmt.Printf("sasageyuo %s \n", conn.Name)
				ls.AddConnection(&conn)
			}
		}
	}()

}

func (cs *ConnectionStack) Push(conn Connection) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	cs.connections = append(cs.connections, conn)
}

func (cs *ConnectionStack) Pop() (Connection, bool) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	if len(cs.connections) == 0 {
		return Connection{}, false
	}
	conn := cs.connections[len(cs.connections)-1]
	cs.connections = cs.connections[:len(cs.connections)-1]
	return conn, true
}
