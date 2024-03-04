package lobbyservice

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Connection struct {
	Name string
	Conn *websocket.Conn
}

// =================================
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

	go newLobby.LifeCycle()
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

// ==========================================================
type Lobby struct {
	filled      bool
	mutex       sync.Mutex
	counter     int
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

	//на случай если он уже был там
	for i, curConn := range lo.Connections {
		if conn.Name == curConn.Name {
			lo.Connections = append(lo.Connections[:i], lo.Connections[i+1:]...)
			curConn.Conn.Close()
			break
		}
	}
	//а теперь добавляем в слайс новый конект
	lo.Connections = append(lo.Connections, conn)
	log.Printf("len of lobby %d is %d \n", &lo, len(lo.Connections))

	if len(lo.Connections) == 5 {
		lo.filled = true
		lo.counter = 5
		log.Println("lobby is filled, starting 5 seconds count")
		// Очистка структуры после написания "done"
	}

	lo.NoticePlayers("updated count players")
}

func (lo *Lobby) LifeCycle() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			for _, lobbyConnection := range lo.Connections {
				err := lobbyConnection.Conn.WriteMessage(websocket.PingMessage, nil)
				if err != nil {
					log.Printf("Юзер %s не пингуется сервером и дропнут из лобби %d \n", lobbyConnection.Name, &lo)
					lo.RemoveConnection(lobbyConnection)
				}
			}
			if len(lo.Connections) < 5 && lo.filled {
				log.Printf("лобби %d было полное и вело 5 сек отсчет, но кто-то отвалился \n", &lo)
				lo.filled = false
			}
			if lo.filled {
				log.Printf("counter in lobby %d is succesfull %d \n", &lo, lo.counter)
				lo.counter -= 1
				if lo.counter < 0 {
					log.Printf("counter in lobby %d is completed! \n", &lo)
					cancel()
				}
			}
		case <-ctx.Done():
			log.Printf("context in lobby %d is done! \n", &lo)
			return
		default:

		}
	}
}

func (lo *Lobby) RemoveConnection(conn *Connection) {
	lo.mutex.Lock()
	defer lo.mutex.Unlock()

	for i, curConn := range lo.Connections {
		if conn.Name == curConn.Name {
			lo.Connections = append(lo.Connections[:i], lo.Connections[i+1:]...)
			curConn.Conn.Close()
			break
		}
	}
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
