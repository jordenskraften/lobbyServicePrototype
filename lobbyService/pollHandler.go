package lobbyservice

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	longPollTimeout = 10 * time.Second
)

type PollResponse struct {
	Message     string `json:"message"`
	RequestFrom string `json:"request_from"`
	BearerToken string `json:"bearer_token,omitempty"`
}
type ErrorResponse struct {
	Message string `json:"error"`
}

func PollHandler(connectionStack *ConnectionStack) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()

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

		// Получаем IP адрес клиента
		ip := r.RemoteAddr
		// Формируем ответ для клиента
		response := PollResponse{
			Message:     "Good response from server!",
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

		conn := Connection{
			Name: bearerToken,
		}
		connectionStack.Push(conn)

		go func(ctx context.Context) {
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			go func(ctx context.Context) {
				ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
				defer cancel()
				<-ctx.Done()
				fmt.Println("Context 2 is done")
			}(ctx)
			<-ctx.Done()
			fmt.Println("Context 1 is done")
		}(ctx)
	}
}

/*
клиент подключился выплюнул токен
дедлайн операции 10 сек от контекста

1. окончание дедлайна? значит от сервака нет сообщений - рефреш
2. сервак хочет послать клиентам в лобби данные
	выслал данные и в ответе флаг рефреш
	состояние клиентов в лобби сохраняется
	если они за 10 сек не переподключились отправив месейдж "рефреш", то удаляем их из лобби
		оповещаем всех клиентов о новом состоянии лобби где стало игроком меньше
3. лобби набилось, оповещаем клиентов в лобби спец.месейджем
	начинается таймер 5 секунд каждую секунду
	так же у клиентов запускается скрипт чтобы каждую секунду шел шорт полл месейджами
	этот прозвон из 5 секунд кончился? окей, инфа о лобби сохраняется в бд
	всем клиентам лобби дается последний меседж с инфой
	далее они пиздуют к клиенту на шарпы


	11. дедлайн лонг пола прошел
	нарн включится горутина или чет похожее
	которая будет ждать в течении 5 сек меседж от этого клиента что он реконектнулся
	его не последовало? ну пока тогда

*/
