package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

var friends []string
var local string
var end chan bool

var ready2listen chan bool
var decisions map[string]string
var cont int

const difficulty = 1

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

// Blockchain is a series of validated Blocks
var Blockchain []Block

// Message takes incoming JSON payload for writing heart rate
type Message struct {
	Edad                      int
	Sexo                      int
	Region                    int
	Viaje                     int
	InsuficienciaRespiratoria int
	Neumonia                  int
	Infectado                 bool
	Riesgo                    float64
}

var mutex = &sync.Mutex{}

func main() {
	//to connect with other nodes
	ready2listen = make(chan bool)
	end = make(chan bool)
	local = os.Args[1]
	decisions = make(map[string]string)
	go serv()
	fmt.Printf(strconv.Itoa(len(os.Args)))
	if len(os.Args) == 3 {
		fmt.Printf("Tengo Amigos")
		remote := os.Args[2]
		cont = 1
		friends = append(friends, os.Args[2])
		send(remote, "hello", []string{}, Block{})
	} else {
		friends = append(friends, os.Args[1])
		addToConsensus(local, "add", []string{local})
	}

	//end to connecto others nodes

	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		t := time.Now()
		genesisBlock := Block{}
		var P = Person{0, 1, 1, 1, 1, 1, true, 0.0}
		genesisBlock = Block{0, t.String(), P, calculateHash(genesisBlock), "", difficulty, ""}
		spew.Dump(genesisBlock)

		mutex.Lock()
		Blockchain = append(Blockchain, genesisBlock)
		mutex.Unlock()
	}()
	log.Fatal(run())
	<-end
	fmt.Println(local, "time to die")
}

// web server
func run() error {
	mux := makeMuxRouter()
	httpPort := os.Getenv("PORT")
	log.Println("HTTP Server Listening on port :", httpPort)
	s := &http.Server{
		Addr:           ":" + httpPort,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := s.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

// create handlers
func makeMuxRouter() http.Handler {
	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/", handleGetBlockchain).Methods("GET")
	muxRouter.HandleFunc("/", handleWriteBlock).Methods("POST")
	return muxRouter
}

// write blockchain when we receive an http request
func handleGetBlockchain(w http.ResponseWriter, r *http.Request) {
	bytes, err := json.MarshalIndent(Blockchain, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	io.WriteString(w, string(bytes))
}

// takes JSON payload as an input for heart rate (BPM)
func handleWriteBlock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var m Message
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		respondWithJSON(w, r, http.StatusBadRequest, r.Body)
		return
	}
	defer r.Body.Close()
	var p Person = Person{m.Edad, m.Sexo, m.Region, m.Viaje, m.InsuficienciaRespiratoria, m.Neumonia, m.Infectado, m.Riesgo}
	fmt.Print(p.Edad)
	fmt.Print(p.Sexo)
	//ensure atomicity when creating new block
	mutex.Lock()
	newBlock := generateBlock(Blockchain[len(Blockchain)-1], p)
	mutex.Unlock()

	/*	if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]) {
		Blockchain = append(Blockchain, newBlock)
		spew.Dump(Blockchain)
	}*/

	respondWithJSON(w, r, http.StatusCreated, newBlock)

	send("localhost:9000", "test", []string{}, newBlock)

}

func respondWithJSON(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	response, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("HTTP 500: Internal Server Error"))
		return
	}
	w.WriteHeader(code)
	w.Write(response)
}

// make sure block is valid by checking index, and comparing the hash of the previous block
func isBlockValid(newBlock, oldBlock Block) bool {
	if oldBlock.Index+1 != newBlock.Index {
		return false
	}

	if oldBlock.Hash != newBlock.PrevHash {
		return false
	}

	if calculateHash(newBlock) != newBlock.Hash {
		return false
	}

	return true
}

// SHA256 hasing
func calculateHash(block Block) string {

	if len(Blockchain) == 0 {

		record := "10:10:10" + strconv.Itoa(block.Persona.Edad) + strconv.Itoa(block.Persona.Sexo) + strconv.Itoa(block.Persona.Region) + strconv.Itoa(block.Persona.Viaje) + strconv.Itoa(block.Persona.Viaje) + strconv.Itoa(block.Persona.InsuficienciaRespiratoria) + strconv.Itoa(block.Persona.Neumonia) + strconv.FormatBool(block.Persona.Infectado) + strconv.FormatFloat(block.Persona.Riesgo, 'E', -1, 64) + block.PrevHash + block.Nonce
		h := sha256.New()
		h.Write([]byte(record))
		hashed := h.Sum(nil)
		return hex.EncodeToString(hashed)
	} else {
		record := block.Timestamp + strconv.Itoa(block.Persona.Edad) + strconv.Itoa(block.Persona.Sexo) + strconv.Itoa(block.Persona.Region) + strconv.Itoa(block.Persona.Viaje) + strconv.Itoa(block.Persona.Viaje) + strconv.Itoa(block.Persona.InsuficienciaRespiratoria) + strconv.Itoa(block.Persona.Neumonia) + strconv.FormatBool(block.Persona.Infectado) + strconv.FormatFloat(block.Persona.Riesgo, 'E', -1, 64) + block.PrevHash + block.Nonce
		h := sha256.New()
		h.Write([]byte(record))
		hashed := h.Sum(nil)
		return hex.EncodeToString(hashed)
	}
}

// create a new block using previous block's hash
func generateBlock(oldBlock Block, p Person) Block {
	var newBlock Block

	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.Persona = p
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Difficulty = difficulty

	for i := 0; ; i++ {
		hex := fmt.Sprintf("%x", i)
		newBlock.Nonce = hex
		if !isHashValid(calculateHash(newBlock), newBlock.Difficulty) {
			fmt.Println(calculateHash(newBlock), " do more work!")
			time.Sleep(time.Second)
			continue
		} else {
			fmt.Println(calculateHash(newBlock), " work done!")
			newBlock.Hash = calculateHash(newBlock)
			break
		}

	}
	return newBlock
}

func isHashValid(hash string, difficulty int) bool {
	prefix := strings.Repeat("0", difficulty)
	return strings.HasPrefix(hash, prefix)
}

//to connecto with other nodes

func serv() {
	fmt.Print("Sirvo de algo")
	// fmt.Println("(", local, ")")
	ln, _ := net.Listen("tcp", local)
	defer ln.Close()
	for {
		conn, _ := ln.Accept()
		go handle(conn)
	}
}
func handle(conn net.Conn) {
	fmt.Print("Entre al handle")
	defer conn.Close()
	dec := json.NewDecoder(conn)
	var msg Msg
	if err := dec.Decode(&msg); err == nil {
		switch msg.Command {
		case "hello":
			resp := Msg{"hey", local, friends, Block{0, "", Person{0, 0, 0, 0, 0, 0, true, 0.0}, "", "", 0, ""}}
			enc := json.NewEncoder(conn)
			if err := enc.Encode(&resp); err == nil {
				for _, friend := range friends {

					// fmt.Println(local, friend, "meet", msg.Hostname)
					send(friend, "meet new friend", []string{msg.Hostname}, Block{0, "", Person{0, 0, 0, 0, 0, 0, true, 0.0}, "", "", 0, ""})
					addToConsensus(friend, "add", []string{msg.Hostname})
				}
			}
			done := make(chan bool, 1)
			var ag = true
			for _, friend := range friends {
				if friend == msg.Hostname {
					ag = false
				}
			}
			done <- true
			if ag {
				<-done
				fmt.Println("Agregando 2", msg.Hostname)
				friends = append(friends, msg.Hostname)
			}
			// fmt.Println(local, "updated list", friends)
		case "meet new friend":

			for _, f := range msg.List {
				var ag2 = true
				if f == local {
					ag2 = false
				}
				done := make(chan bool, 1)
				for _, friend := range friends {
					if friend == f || f == local {
						ag2 = false
						break
					}

				}
				done <- true
				<-done
				if ag2 {
					fmt.Println("Agregando ", f)
					friends = append(friends, msg.List...)
				}
			}

			addToConsensus(local, "add", msg.List)
			// fmt.Println(local, "new friend", msg.List)

		case "test consensus":

			b := msg.Block
			fmt.Println("Haciendo Test")
			fmt.Println(msg.Block)
			if isBlockValid(b, Blockchain[len(Blockchain)-1]) {
				fmt.Println("Bloque valido")
				decisions[local] = "aceptar"
			} else {
				fmt.Println("Bloque no valido")
				decisions[local] = "ignorar"
			}
			fmt.Println(local, decisions[local])
			for _, friend := range friends {
				send(friend, "decision", []string{decisions[local]}, b)
			}
			ready2listen <- true
		case "decision":

			<-ready2listen
			cont = cont + 1
			fmt.Println("valor inicial " + strconv.Itoa(cont))
			fmt.Println("Decisiones " + decisions[local])
			fmt.Println("Lista: " + msg.List[0])
			decisions[msg.Hostname] = msg.List[0]

			b := msg.Block
			fmt.Print(friends)
			fmt.Println("tamaño: " + strconv.Itoa(len(friends)))
			fmt.Println("contador" + strconv.Itoa(cont))
			if cont == len(friends) {
				contAtack := 0
				contFallb := 0

				for _, decision := range decisions {
					if decision == "aceptar" {
						contAtack++
					} else {
						contFallb++
					}
				}
				fmt.Println("Atacar " + strconv.Itoa(contAtack))
				fmt.Println("Retirarse" + strconv.Itoa(contFallb))
				if contAtack < contFallb {
					//ignorar
					fmt.Println(local, "RETIRADA!!")
				} else {
					//aceptar
					Blockchain = append(Blockchain, b)
					spew.Dump(Blockchain)
				}
				if len(os.Args) == 3 {
					cont = 0
				} else {
					cont = 1
				}
				end <- true
			} else {
				ready2listen <- true
			}

		case "finish":
			end <- true
		}
	}
}
func send(remote, command string, list []string, block Block) {
	fmt.Println("send")
	conn, _ := net.Dial("tcp", remote)
	defer conn.Close()
	msg := Msg{command, local, list, block}
	enc := json.NewEncoder(conn)
	if err := enc.Encode(&msg); err == nil {
		// fmt.Println(local, "sent", msg)
		if command == "hello" {
			dec := json.NewDecoder(conn)
			var resp Msg
			if err := dec.Decode(&resp); err == nil {
				friends = append(friends, resp.List...)
				// fmt.Println(local, "recibí", resp.List)
			}
		}
	}
}
func addToConsensus(remote, command string, list []string) {
	send("localhost:9000", "add", list, Block{0, "", Person{0, 0, 0, 0, 0, 0, true, 0.0}, "", "", 0, ""})
}
