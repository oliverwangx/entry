package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:8989")
	if err != nil {
		log.Fatalln(err)
	}
	defer listener.Close()

	for {
		con, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go handleClientRequest(con)
	}
}

func handleClientRequest(con net.Conn) {
	defer con.Close()

	clientReader := bufio.NewReader(con)
	userMap := map[string]string{"wxy": "950822"}
	status := "Failed"
	for {
		clientRequest, err := clientReader.ReadString('\n')
		switch err {
		case nil:
			clientRequest := strings.TrimSpace(clientRequest)
			// fmt.Println("clientRequest: ", clientRequest)
			sec := map[string]string{}
			if err := json.Unmarshal([]byte(clientRequest), &sec); err != nil {
				panic(err)
			}
			fmt.Println("map is the ", sec["password"], sec["username"])
			if val, ok := userMap[sec["username"]]; ok && val == sec["password"] {
				log.Println("user exits")
				status = "success"
			} else {
				log.Println("user log in failed")

			}

			if clientRequest == ":QUIT" {
				log.Println("client requested server to close the connection so closing")
				return
			}
		case io.EOF:
			log.Println("client closed the connection by terminating the process")
			return
		default:
			log.Printf("error: %v\n", err)
			return
		}

		if _, err = con.Write(createResponse(status)); err != nil {
			log.Printf("failed to respond to client: %v\n", err)
		}
	}
}

func createResponse(status string) []byte {
	return []byte(
		fmt.Sprintf(`{"code":%d,"data":{"id":"%s","created_at":"%s", "status": "%s"}}`+"\n",
			http.StatusOK,
			uuid.New().String(),
			time.Now().UTC().Format(time.RFC3339),
			status,
		),
	)
}
