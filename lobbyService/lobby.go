package lobbyservice

import (
	"fmt"
	"sync"
)

type Lobby struct {
	connections []Connection
	mutex       sync.Mutex
}

func (us *Lobby) AddConnection(conn *Connection) {
	us.mutex.Lock()
	defer us.mutex.Unlock()
	us.connections = append(us.connections, *conn)
	fmt.Printf("len %d \n", len(us.connections))
	if len(us.connections) == 5 {
		fmt.Println("done")
		// Очистка структуры после написания "done"
		us.connections = nil
	}
}
