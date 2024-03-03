package main

import (
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
	lobby := lobbyservice.Lobby{
		Ch: make(chan string),
	}
	connectionStack.ObserveLobby(&lobby)

	//давай потестируем лобби
	go lobbyTesting(&lobby)

	for {

	}
	//для начала роутер на эндпоинт
	r := chi.NewRouter()
	r.Get("/poll", lobbyservice.PollHandler(&connectionStack))

	http.ListenAndServe(":8080", r)

}

func lobbyTesting(lo *lobbyservice.Lobby) {
	conn1 := &lobbyservice.Connection{
		Name:    "111",
		Writter: nil,
	}
	conn2 := &lobbyservice.Connection{
		Name:    "222",
		Writter: nil,
	}
	conn3 := &lobbyservice.Connection{
		Name:    "333",
		Writter: nil,
	}

	fmt.Println("one conn")
	lo.AddConnection(conn1)
	time.Sleep(2 * time.Second)
	lo.Ch <- conn1.Name
	time.Sleep(6 * time.Second) // Пауза для обработки сообщения
	fmt.Println("=======================")

	fmt.Println("two conn")
	lo.AddConnection(conn2)
	time.Sleep(2 * time.Second)
	lo.Ch <- conn1.Name
	lo.Ch <- conn2.Name
	time.Sleep(6 * time.Second) // Пауза для обработки сообщений
	fmt.Println("=======================")

	fmt.Println("three conn")
	lo.AddConnection(conn3)
	time.Sleep(2 * time.Second)
	lo.Ch <- conn1.Name
	lo.Ch <- conn2.Name
	lo.Ch <- conn3.Name
	time.Sleep(6 * time.Second) // Пауза для обработки сообщений
	fmt.Println("=======================")
}
