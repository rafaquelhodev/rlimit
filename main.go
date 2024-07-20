package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/rafaquelhodev/rlimit/internal/tokens"
)

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

		options := strings.Split(temp, " ")

		resp := ""
		if len(options) < 3 {
			resp = "INVALID: reqID maxTokens refillRate are required\n"
		} else {
			reqID := options[0]

			maxToken, err := strconv.ParseInt(options[1], 10, 64)
			if err != nil {
				resp = fmt.Sprintf("%s: INVALID ERROR: maxTokens should be an integer\n", reqID)
			}

			refillRate, err := strconv.ParseInt(options[2], 10, 64)
			if err != nil {
				resp = fmt.Sprintf("%s: INVALID ERROR: refillRate should be an integer\n", reqID)
			}

			if resp == "" {
				tm.WaitAvailable(reqID, maxToken, refillRate)

				resp = fmt.Sprintf("%s: AVAILABLE\n", reqID)
			}

		}

		c.Write([]byte(string(resp)))
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
