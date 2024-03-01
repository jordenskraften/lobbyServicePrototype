package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"longPoll/lobbyservice"

	"github.com/go-chi/chi"
)

// -------------------------------
func main() {
	fmt.Println("service is started")

	//есть пул подключений
	//есть лобби которое берет кого-то из пула
	//есть механизм который проверяет надо ли спавнить лобби или одобрить его когда оно заполнено

	connectionStack := lobbyservice.ConnectionStack{}
	lobbyStruct := lobbyservice.Lobby{}

	// Извлечение соединений из стека и добавление их в структуру
	go func() {
		for {
			conn, ok := connectionStack.Pop()
			if ok {
				fmt.Printf("sasageyuo %s \n", conn.Name)
				lobbyStruct.AddConnection(&conn)
			}
		}
	}()

	//для начала роутер на эндпоинт
	r := chi.NewRouter()
	r.Get("/poll", pollHandler(&connectionStack))

	http.ListenAndServe(":8080", r)

}

type PollResponse struct {
	Message     string `json:"message"`
	RequestFrom string `json:"request_from"`
	BearerToken string `json:"bearer_token,omitempty"`
}
type ErrorResponse struct {
	Message string `json:"error"`
}

func pollHandler(connectionStack *lobbyservice.ConnectionStack) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		bearerToken := r.Header.Get("Authorization")
		// Получаем значение токена Bearer из заголовка запроса
		// Если был предоставлен токен Bearer, добавляем его в ответ
		if bearerToken == "" {
			// Формируем ответ с ошибкой доступа
			errorResponse := ErrorResponse{
				Message: "Unauthorized: Bearer token is required",
			}
			// Кодируем структуру в JSON
			jsonResponse, _ := json.Marshal(errorResponse)

			// Устанавливаем статус код 401 (Unauthorized)
			w.WriteHeader(http.StatusUnauthorized)
			// Устанавливаем заголовок Content-Type на application/json
			w.Header().Set("Content-Type", "application/json")

			// Отправляем ответ клиенту
			w.Write(jsonResponse)
			return
		}

		// Задержка в 2 секунды для демонстрации длительного соединения
		time.Sleep(2 * time.Second)

		// Получаем IP адрес клиента
		ip := r.RemoteAddr
		// Формируем ответ для клиента
		response := PollResponse{
			Message:     "Long Poll time is Over!",
			RequestFrom: ip,
			BearerToken: bearerToken,
		}

		// Кодируем структуру в JSON
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Устанавливаем заголовок Content-Type на application/json
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResponse)

		conn := lobbyservice.Connection{
			Name: bearerToken,
		}
		connectionStack.Push(conn)
	}
}

//======================
