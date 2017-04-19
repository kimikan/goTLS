package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func main() {

	if len(os.Args) != 3 {
		fmt.Println("Usage: tls [address] [port]")
		return
	}

	var addr = os.Args[1] + ":" + os.Args[2]

	CAPool := x509.NewCertPool()
	severCert, err := ioutil.ReadFile("ServerCert.pem")
	if err != nil {
		log.Fatal("Could not load server certificate! ", err)
	}
	CAPool.AppendCertsFromPEM(severCert)

	cert, err := tls.LoadX509KeyPair("ClientCert.pem", "ClientKey.pem")
	config := &tls.Config{
		RootCAs:      CAPool,
		Certificates: []tls.Certificate{cert},
	}
	/*conf := &tls.Config{
	InsecureSkipVerify: true,
	}*/

	conn, err := tls.Dial("tcp", addr, config)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)
	go func() {
		buf2 := make([]byte, 100)
		for {
			len, err2 := conn.Read(buf2)
			if err2 != nil {
				return
			}
			println(string(buf2[:len]))
		}
	}()

	for {
		data, _, _ := reader.ReadLine()
		command := string(data)
		conn.Write([]byte(command))
	}
}
