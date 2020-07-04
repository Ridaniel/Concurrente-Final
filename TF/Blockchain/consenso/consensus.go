package main

import (
	"encoding/json"
	"fmt"
	"net"
)

var unique = make(map[string]bool, 10000)
var us = make([]string, len(unique))
var end chan bool

var ready2listen chan bool

type Msg struct {
	Command  string
	Hostname string
	List     []string
	Block    Block
}

// Block represents each 'item' in the blockchain
type Block struct {
	Index      int
	Timestamp  string
	Persona    Person
	Hash       string
	PrevHash   string
	Difficulty int
	Nonce      string
}

//Person Representa una persona
type Person struct {
	Edad                      int
	Sexo                      int
	Region                    int
	Viaje                     int
	InsuficienciaRespiratoria int
	Neumonia                  int
	Infectado                 bool
	Riesgo                    float64
}

func serv() {
	// fmt.Println("(", local, ")")
	fmt.Println("entre")
	ln, _ := net.Listen("tcp", "localhost:9000")
	defer ln.Close()

	for {
		fmt.Println("escuchando")
		conn, _ := ln.Accept()
		go handle(conn)
	}
}

func handle(conn net.Conn) {

	defer conn.Close()
	dec := json.NewDecoder(conn)
	var msg Msg
	fmt.Print("Se conectaron")
	if err := dec.Decode(&msg); err == nil {
		switch msg.Command {
		case "add":

			for _, elem := range msg.List {
                if len(elem) != 0 {
                    if !unique[elem] {
                        us = append(us, elem)
                        unique[elem] = true
                    }
                }
            }
			fmt.Println("Nodo añadido")
			fmt.Println(us)
			// fmt.Println(local, "updated list", friends)
			ready2listen <- true
		case "test":
			<-ready2listen
			for _, friend := range us {

				send(friend, "test consensus", []string{}, msg.Block)

			}
		}

	}
}

func send(remote, command string, list []string, b Block) {
	conn, _ := net.Dial("tcp", remote)
	defer conn.Close()
	msg := Msg{command, "", list, b}
	enc := json.NewEncoder(conn)
	if err := enc.Encode(&msg); err == nil {
		fmt.Println("sending test consensus to", remote)
	}
}

func main() {

	ready2listen = make(chan bool)
	end = make(chan bool)
	go serv()
	<-end
}
