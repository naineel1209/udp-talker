package bareudp

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

func UdpText() {
	fmt.Println("Hello, World!")

	//connecting to the udp server
	conn, err := net.DialUDP("udp4",
		&net.UDPAddr{ //local server address
			IP:   net.IPv4(127, 0, 0, 1),
			Port: 41235,
			Zone: "",
		},
		&net.UDPAddr{ //remote server address
			IP:   net.IPv4(127, 0, 0, 1),
			Port: 41234,
			Zone: "",
		})

	if err != nil {
		log.Fatalf("error: %v", err)
	}

	defer conn.Close()

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024), 1024) // set the buffer size to 1024 bytes

	for {
		//sending the message
		conn.SetWriteDeadline(time.Now().Add(5 * time.Second)) // set the write deadline to 5 seconds - if the message is not sent within 5 seconds, it will return an error
		var msg string
		fmt.Print("\n Enter message: \n")
		for scanner.Scan() {
			if scanner.Err() != nil {
				log.Fatalf("error: %v", scanner.Err().Error())
			}

			msg = scanner.Text()
			break
		}

		fmt.Printf("Sending: %s\n", msg)

		// msg, err = reader.ReadString('\n')

		if err != nil {
			log.Fatalf("error: %v", err)
		}

		write_n, _, err := conn.WriteMsgUDP([]byte(msg), nil, nil)

		if err != nil {
			log.Fatalf("error: %v", err)
		}

		fmt.Printf("Sent %d bytes\n", write_n)

		//receiving the message
		buffer := make([]byte, 1024)

		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		read_n, _, err := conn.ReadFrom(buffer) // read the message from the server and store it in the buffer variable - if the message is not received within 5 seconds, it will return an error

		if err != nil {
			log.Fatalf("error: %v", err)
		}

		fmt.Printf("Received: %s\n of length: %v bytes", string(buffer[:read_n]), read_n)
	}
}
