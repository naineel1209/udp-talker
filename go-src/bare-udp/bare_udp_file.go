package bareudp

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"
)

type UdpFileStruct struct {
	Filename    string `json:"filename"`
	Filesize    int64  `json:"filesize"`
	Currentbyte int64  `json:"currentbyte"`
	Data        []byte `json:"data"`
}

func UdpFile() {
	conn, err := net.DialUDP("udp4", &net.UDPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 41236,
		Zone: "",
	}, &net.UDPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 41234,
		Zone: "",
	})

	if err != nil {
		log.Fatalf("error: %v", err)
	}

	defer conn.Close()

	cwd, err := os.Getwd()
	fmt.Println(cwd)

	if err != nil {
		log.Fatalf("error: %v", err)
	}

	var filename string
	fmt.Print("Enter the filename: ")
	fmt.Scan(&filename)

	//join the paths
	filePath := filepath.Join(cwd, "files", filename)

	file, err := os.OpenFile(filePath, os.O_RDONLY, 0666) // 666 is the permission for the file - read and write permission

	if err != nil {
		log.Fatalf("error: %v", err)
	}

	defer file.Close()

	//get the file size in bytes
	fileInfo, err := file.Stat()

	if err != nil {
		log.Fatalf("error: %v", err)
	}

	fileSize := fileInfo.Size()

	BUFFER_SIZE := 1024 * 10             //10KB
	conn.SetWriteBuffer(BUFFER_SIZE + 1) // set the write buffer to 1MB

	// read the file in chunks of 1024 bytes
	buffer := make([]byte, BUFFER_SIZE)
	readBytes := int64(0)

	//start the timer
	startTime := time.Now().Unix()

	//send the file in chunks of 1024 bytes
	for readBytes < fileSize {
		read_n, err := file.Read(buffer)

		if err == io.EOF {
			fmt.Println("EOF")
			break
		}

		//read the first n bytes of the buffer
		// this is done to avoid sending the entire buffer of 1024 bytes
		buffer = buffer[:read_n]

		//print the buffer
		// fmt.Printf("Sending: %s\n", string(buffer))

		readBytes += int64(read_n)

		_, err = file.Seek(int64(readBytes), 0)

		if err != nil {
			log.Fatalf("error: %v", err)
		}

		msgFile := UdpFileStruct{
			Filename:    fileInfo.Name(),
			Filesize:    fileSize,
			Currentbyte: readBytes,
			Data:        buffer,
		}

		dataBytes, err := json.MarshalIndent(msgFile, "", "  ")

		if err != nil {
			log.Fatalf("error: %v", err)
		}
		//send the buffer
		write_n, _, err := conn.WriteMsgUDP(dataBytes, nil, nil)

		if err != nil {
			log.Fatalf("error: %v", err)
		}

		fmt.Printf("Sent %d bytes\n", write_n)
		fmt.Println("=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-")

		//Wait for the server to acknowledge the message
		bufferAck := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(bufferAck) //this is a blocking call - it will wait for the server to send the ACK

		if err != nil {
			log.Fatalf("error: %v", err)
		}

		if string(bufferAck[:n]) == "ACK" {
			fmt.Println("ACK received")
		} else {
			fmt.Println("ACK not received")
		}
	}

	fmt.Printf("File sent successfully\n")
	fmt.Printf("Total bytes sent: %d\n", readBytes)
	fmt.Printf("Time taken: %v seconds\n", time.Now().Unix()-int64(startTime)) //convert the time to minutes in form of  - 1m30s
}
