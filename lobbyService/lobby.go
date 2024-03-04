package lobbyservice

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

type Connection struct {
	Name string
	Conn *websocket.Conn
}

type LobbyPull struct {
	lobbies []*Lobby
	mutex   sync.Mutex
}

func NewLobbyPull() *LobbyPull {
	return &LobbyPull{
		lobbies: make([]*Lobby, 0),
		mutex:   sync.Mutex{},
	}
}

func (lp *LobbyPull) CreateLobby() *Lobby {
	newLobby := &Lobby{
		filled:      false,
		mutex:       sync.Mutex{},
		Connections: make([]*Connection, 0, 5),
	}
	lp.lobbies = append(lp.lobbies, newLobby)
	return newLobby

}

func (lp *LobbyPull) AddConnectionToLobby(conn *Connection) {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()

	var freeLobby *Lobby

	if len(lp.lobbies) == 0 {
		freeLobby = lp.CreateLobby()
		fmt.Println("creating first lobby in lobby pull")
	} else {
		for _, lobby := range lp.lobbies {
			if lobby.IsFilled() == false {
				freeLobby = lobby
				fmt.Println("found free lobby in lobby pull")
			}
		}
		if freeLobby == nil {
			freeLobby = lp.CreateLobby()
			fmt.Println("no free lobby in lobby pull, creating new")
		}
	}
	freeLobby.AddConnection(conn)
	fmt.Println(lp.lobbies)
}

type Lobby struct {
	filled      bool
	mutex       sync.Mutex
	Connections []*Connection
}

//думаю дать лобби метод который запустит горутину с тикером и кансел контекстом
//эта горутина будет пинговать плееров в лобби и в целом "жить"

func (lo *Lobby) AddConnection(conn *Connection) {
	lo.mutex.Lock()
	defer lo.mutex.Unlock()

	if lo.Connections == nil {
		lo.Connections = make([]*Connection, 0, 5)
	}

	lo.Connections = append(lo.Connections, conn)
	fmt.Printf("len %d \n", len(lo.Connections))

	if len(lo.Connections) == 5 {
		lo.filled = true
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
}

// лобби должно уметь дожидаться реконект игроков после месейджа
func (lo *Lobby) LeaversCheck() {
}

//
