package main

import (
	"encoding/json"
	"log"
	lobbyservice "longPoll/lobbyService"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type AuthorizationMessage struct {
	Authorization string `json:"authorization"`
}

func handleWebSocket(lp *lobbyservice.LobbyPull) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Ошибка при обновлении соединения веб-сокета:", err)
			return
		}
		defer conn.Close()

		// Читаем данные, отправленные клиентом сразу после открытия соединения
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Ошибка чтения сообщения:", err)
			return
		}
		log.Printf("Получены данные от клиента при открытии соединения: %s\n", message)

		var authMessage AuthorizationMessage
		err = json.Unmarshal([]byte(message), &authMessage)
		if err != nil {
			log.Println("Ошибка декодирования JSON:", err)
			return
		}

		time.Sleep(1 * time.Second) // Отправляем пинг каждые 5 секунд
		err = conn.WriteMessage(websocket.PingMessage, nil)
		if err != nil {
			log.Println("Ошибка отправки пинга:", err)
			//окей эта хуйня пашет и проверяет дисконектнулся ли кто-то или нет
			return
		}
		time.Sleep(1 * time.Second) // Отправляем пинг каждые 5 секунд

		Connection := lobbyservice.Connection{
			Name: authMessage.Authorization,
			Conn: conn,
		}
		lp.AddConnectionToLobby(&Connection)
	}
}

func main() {

	lobbyPull := lobbyservice.NewLobbyPull()

	http.HandleFunc("/ws", handleWebSocket(lobbyPull))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
