package lobbyservice

import (
	"context"
	"encoding/json"
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
		lobbyPull:   lp,
		mutex:       sync.Mutex{},
		Connections: make([]*Connection, 0, 5),
	}
	lp.lobbies = append(lp.lobbies, newLobby)

	go newLobby.LifeCycle()
	return newLobby

}

func (lp *LobbyPull) RemoveLobby(lobby *Lobby) {
	for i, curLobby := range lp.lobbies {
		if lobby == curLobby {
			log.Printf("lobby pull before deletion lobby %v", lp.lobbies)
			lp.lobbies = append(lp.lobbies[:i], lp.lobbies[i+1:]...)
			log.Printf("lobby %d is deleted from lobby pull", &lobby)
			log.Printf("lobby pull after deletion lobby %v", lp.lobbies)
			break
		}
	}
}

func (lp *LobbyPull) AddConnectionToLobby(conn *Connection) {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()

	var freeLobby *Lobby

	log.Printf("lobby pull before creating new lobby %v", lp.lobbies)
	if len(lp.lobbies) == 0 {
		freeLobby = lp.CreateLobby()
		log.Println("creating first lobby in lobby pull")
	} else {
		for _, lobby := range lp.lobbies {
			if lobby.IsFilled() == false {
				freeLobby = lobby
				log.Println("found free lobby in lobby pull")
			}
		}
		if freeLobby == nil {
			freeLobby = lp.CreateLobby()
			log.Println("no free lobby in lobby pull, creating new")
		}
	}
	freeLobby.AddConnection(conn)
	log.Printf("lobby pull after creating new lobby %v", lp.lobbies)
}

// ==========================================================
type Lobby struct {
	filled      bool
	mutex       sync.Mutex
	finalTimer  int
	lobbyPull   *LobbyPull
	Connections []*Connection
}
type LobbyCountMessage struct {
	PlayersCount int `json:"players_count"`
}
type LobbyTokenMessage struct {
	LobbyToken string `json:"lobby_token"`
}
type LobbyTimerMessage struct {
	FinalTimer int `json:"final_timer"`
}
type LobbySpecialMessage struct {
	SpecialMessage string `json:"special_message"`
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
		lo.finalTimer = 5
		log.Println("lobby is filled, starting 5 seconds count")
		// Очистка структуры после написания "done"
	}
	msg := &LobbyCountMessage{
		PlayersCount: len(lo.Connections),
	}
	lo.NoticePlayers(msg)
}

func (lo *Lobby) LifeCycle() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			lo.LeaversCheckAndDrop()
			if len(lo.Connections) < 5 && lo.filled {
				log.Printf("лобби %d было полное и вело 5 сек отсчет, но кто-то отвалился \n", &lo)
				lo.filled = false

				msg := &LobbySpecialMessage{
					SpecialMessage: "start is interrupted, waiting full lobby again",
				}
				lo.NoticePlayers(msg)
			}
			if lo.filled {
				log.Printf("counter in lobby %d is succesfull %d \n", &lo, lo.finalTimer)
				msg := &LobbyTimerMessage{
					FinalTimer: lo.finalTimer,
				}
				lo.NoticePlayers(msg)

				lo.finalTimer -= 1
				if lo.finalTimer < 0 {
					log.Printf("counter in lobby %d is completed! \n", &lo)
					cancel()
				}
			}
		case <-ctx.Done():
			log.Printf("context in lobby %d is done! \n", &lo)
			token := lo.GenerateLobbyToken()

			msg := &LobbyTokenMessage{
				LobbyToken: token,
			}
			lo.NoticePlayers(msg)
			lo.FinalizeLobby()
			return
		default:

		}
	}
}

func (lo *Lobby) FinalizeLobby() {

	for _, lobbyConnection := range lo.Connections {
		lobbyConnection.Conn.Close()
	}
	lo.Connections = nil
	lo.lobbyPull.RemoveLobby(lo)
	log.Printf("lobby %d is cleared and destroyed", &lo)
}

func (lo *Lobby) GenerateLobbyToken() string {
	token := fmt.Sprintf("", &lo)
	return token
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
	msg := &LobbyCountMessage{
		PlayersCount: len(lo.Connections),
	}
	lo.NoticePlayers(msg)
}

func (lo *Lobby) IsFilled() bool {
	return lo.filled
}

// лобби должно уметь оповещать всех игроков о событии

func (lo *Lobby) NoticePlayers(message interface{}) {
	// Преобразование message в JSON
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		log.Printf("ошибка преобразования сообщения в JSON: %v", err)
		return
	}

	// Отправка JSON-сообщения каждому подключенному игроку
	for _, playerConn := range lo.Connections {
		err := playerConn.Conn.WriteMessage(websocket.TextMessage, jsonMessage)
		if err != nil {
			log.Printf("не удалось отправить сообщение игроку %s: %v", playerConn.Name, err)
			continue // Продолжаем отправку сообщений другим игрокам
		}
		log.Printf("отправлено сообщение игроку %s: %s", playerConn.Name, jsonMessage)
	}
}

func (lo *Lobby) LeaversCheckAndDrop() {
	for _, lobbyConnection := range lo.Connections {
		err := lobbyConnection.Conn.WriteMessage(websocket.PingMessage, nil)
		if err != nil {
			log.Printf("Юзер %s не пингуется сервером и дропнут из лобби %d \n", lobbyConnection.Name, &lo)
			lo.RemoveConnection(lobbyConnection)
		}
	}
}

//
