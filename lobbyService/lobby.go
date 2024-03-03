package lobbyservice

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type LobbyPull struct {
	lobbies map[string]*Lobby
	mutex   sync.Mutex
}

type Lobby struct {
	filled      bool
	mutex       sync.Mutex
	LobbyId     string
	Ch          chan string
	Connections map[string]*Connection
}

const (
	leaversTimer = time.Second * 3
)

func (lo *Lobby) AddConnection(conn *Connection) {
	lo.mutex.Lock()
	defer lo.mutex.Unlock()

	if lo.Connections == nil {
		lo.Connections = make(map[string]*Connection)
	}

	lo.Connections[conn.Name] = conn
	fmt.Printf("len %d \n", len(lo.Connections))

	if len(lo.Connections) == 5 {
		fmt.Println("lobby is done")
		// Очистка структуры после написания "done"
	}

	lo.NoticePlayers("updated count players")
}

func (lo *Lobby) IsFilled() bool {
	return lo.filled
}

// лобби должно уметь оповещать всех игроков о событии
func (lo *Lobby) NoticePlayers(message string) {
	for _, a := range lo.Connections {
		fmt.Println("noticed player", a.Name)
	}

	go lo.LeaversCheck()
}

// лобби должно уметь дожидаться реконект игроков после месейджа
func (lo *Lobby) LeaversCheck() {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, leaversTimer)
	defer cancel()

	survivors := make(map[string]bool)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				fmt.Println("done ctx")
				return
			case str := <-lo.Ch:
				fmt.Println("got chan msg", str)
				survivors[str] = true
			default:
			}
		}
	}()

	wg.Wait()

	prevLen := len(lo.Connections)

	lo.mutex.Lock()
	for connID := range lo.Connections {
		if _, exists := survivors[connID]; !exists {
			fmt.Println("not found in survivors", connID)
			delete(lo.Connections, connID)
		}
	}
	lo.mutex.Unlock()

	newLen := len(lo.Connections)

	fmt.Printf("prev len %d, new len %d \n", prevLen, newLen)
}

//
