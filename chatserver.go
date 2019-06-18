package main

import (
    "bufio"
    "fmt"
    "log"
    "net"
    "time"
)

//const timeout = time.Second * 10

type client chan<- string // an outgoing message channel

var (
    entering = make(chan client)
    leaving  = make(chan client)
    messages = make(chan string) // all incoming client messages
)

func broadcaster() {
    clients := make(map[client]bool) // all connected clients
    for {
        select {
        case msg := <-messages:
            for cli := range clients {
                cli <- msg
            }

        case cli := <-entering:
            clients[cli] = true

        case cli := <-leaving:
            delete(clients, cli)
            close(cli)
        }
    }
}

func handleConn(conn net.Conn) {
    starttime := time.Now()
    ch := make(chan string) // outgoing client messages
    go clientWriter(conn, ch)
    recentmessage := make(chan time.Time)
    who := conn.RemoteAddr().String()
    go func(t time.Time) {
        for {
            select {
            case t = <-recentmessage:
            default:
            }
            if time.Now().After(t.Add(60 * time.Second)) {
                conn.Close()
                break
            }
        }
    }(starttime)
    ch <- "You are " + who
    messages <- who + " has joined"
    entering <- ch

    input := bufio.NewScanner(conn)
    for input.Scan() {
        messages <- who + ": " + input.Text()
        recentmessage <- time.Now()
    }

    leaving <- ch
    messages <- who + " has left"
    conn.Close()
}

func clientWriter(conn net.Conn, ch <-chan string) {
    for msg := range ch {
        fmt.Fprintln(conn, msg)
    }
}


func main() {
    listener, err := net.Listen("tcp", "localhost:8000")
    if err != nil {
        log.Fatal(err)
    }

    go broadcaster()
    for {
        conn, err := listener.Accept()
        if err != nil {
            log.Print(err)
            continue
        }
        go handleConn(conn)
    }

}