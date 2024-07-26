package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/rafaquelhodev/rlimit/internal/tokens"
)

type ExecType map[string](func(tm *tokens.TokenBucketManager, options []string) (string, error))

func executeRequest(tm *tokens.TokenBucketManager, options []string) (string, error) {
	if len(options) == 0 {
		return "", errors.New("please provide options")
	}

	switch options[0] {
	case "INIT":
		if len(options) < 4 {
			return "", errors.New("reqID maxToken refillRate required")
		}

		reqID := options[1]

		maxToken, err := strconv.ParseInt(options[2], 10, 64)
		if err != nil {
			return "", fmt.Errorf("maxTokens should be an integer")
		}

		refillRate, err := strconv.ParseInt(options[3], 10, 64)
		if err != nil {
			return "", fmt.Errorf("refillRate should be an integer")
		}

		tm.CreateTokenBucket(reqID, maxToken, refillRate)

		return fmt.Sprintf("%s: DONE", reqID), nil
	case "CHECK":
		if len(options) < 2 {
			return "", errors.New("reqID is required")
		}

		reqID := options[1]

		err := tm.WaitAvailable(reqID)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%s: AVAILABLE", reqID), nil
	default:
		return "", errors.New("please provide a valid option")
	}
}

func handleConnection(c net.Conn, tm *tokens.TokenBucketManager) {
	fmt.Printf("Serving %s\n", c.RemoteAddr().String())

	for {
		netData, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}

		temp := strings.TrimSpace(string(netData))
		if temp == "STOP" {
			fmt.Printf("Closing connection %s\n", c.RemoteAddr().String())
			break
		}

		fmt.Printf("Received %s from %s\n", temp, c.RemoteAddr().String())

		go func() {
			options := strings.Split(temp, " ")
			resp, err := executeRequest(tm, options)

			if err != nil {
				c.Write([]byte(string("ERROR: " + err.Error() + "\n")))
			} else {
				c.Write([]byte(string(resp + "\n")))
			}
		}()

	}
	c.Close()
}

func startServer(tm *tokens.TokenBucketManager) {
	args := os.Args
	if len(args) == 1 {
		fmt.Println("Please provide a port number")
	}

	port := ":" + args[1]

	l, err := net.Listen("tcp4", port)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		go handleConnection(c, tm)
	}
}

func main() {
	tokenManager := tokens.NewUserTokenBucketManager()
	startServer(tokenManager)

	// var wg sync.WaitGroup

	// bucket := tokenManager.getTokenBucket("test")

	// for i := 0; i < 50; i++ {
	// 	wg.Add(1)
	// 	go func() {
	// 		bucket.waitAvailable()
	// 		defer wg.Done()
	// 	}()
	// }

	// wg.Wait()

}
