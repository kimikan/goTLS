package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

//Item ..
type Item struct {
	IP         string
	Time       string
	Connection string
	Text       string
}

//Logs ...
var Logs = make([]Item, 1)
var lock sync.Mutex

//Post ...
func Post(w http.ResponseWriter, req *http.Request) {
	text := req.FormValue("Text")

	if text == "" {
		w.Write([]byte("Message is empty! "))
		return
	}
	lock.Lock()
	Logs = append(Logs, Item{
		IP:         req.RemoteAddr,
		Time:       time.Now().String(),
		Text:       text,
		Connection: "https",
	})
	lock.Unlock()
	http.Redirect(w, req, "/Hello", http.StatusMovedPermanently)
}

//Hello ..
func Hello(w http.ResponseWriter, req *http.Request) {
	const tpl = `
    <!DOCTYPE html>
    <html>
        <head>
            <meta charset="UTF-8">
            <title>{{.Title}}</title>
        </head>
        <body>
            <p><h2>Message board</h2></p>
            <table border="1" style="table-layout:fixed;word-break: break-all;">
                <tr><th>IP</th><th>Datetime</th><th>Connection</th><th>Message</th></tr>
                
                {{with .}}
                    {{range .Items}}  
                        <tr><td>{{ .IP }}</td><td>{{ .Time }}</td><td>{{ .Connection }}</td><td>{{ .Text }}</td></tr> 
                    {{end}}  
                    
                {{end}}  
            </table>
            <form action="/Post" method="post"> 
                <input type="text" name="Text">
                <input type="submit" value="submit"></br> 
            </form>
        </body>
    </html>`

	t, _ := template.New("webpage").Parse(tpl)
	data := struct {
		Title string
		Items []Item
	}{
		Title: "Test",
		Items: Logs,
	}
	lock.Lock()
	err := t.Execute(w, data)
	lock.Unlock()
	if err != nil {
		w.Write([]byte(err.Error()))
	}
}

//OnConnection ..
func OnConnection(conn net.Conn) {
	defer conn.Close()
	conn.Write([]byte("command: 1. list\n  2. add [text]\n    "))
	var buf [512]byte
	for {

		n, err := conn.Read(buf[0:])

		if err != nil {
			fmt.Println(conn, err)
			break
		}
		cmd := string(buf[:n])
		fmt.Println(cmd)
		if cmd == "list" {
			for _, value := range Logs {
				conn.Write([]byte(fmt.Sprintf("[%s] [%s] [%s] [%s]\n", value.IP, value.Time, value.Connection, value.Text)))
			}
		} else if strings.HasPrefix(cmd, "add ") {
			lock.Lock()
			Logs = append(Logs, Item{
				IP:         conn.RemoteAddr().String(),
				Time:       time.Now().String(),
				Text:       cmd[4:],
				Connection: "tls",
			})
			lock.Unlock()
		}
	}
}

//TLS ..
func TLS() {

	clientCertsPool := x509.NewCertPool()

	clientCert, err := ioutil.ReadFile("ClientCert.pem")
	if err != nil {
		log.Fatal("Could not load server certificate!")
	}
	if !clientCertsPool.AppendCertsFromPEM(clientCert) {
		log.Fatal("Can not append client cert! ")
	}

	cert, err := tls.LoadX509KeyPair("ServerCert.pem", "ServerKey.pem")
	if err != nil {
		log.Fatal("Error loading certificate. ", err)
	}

	tlsCfg := &tls.Config{
		Certificates:             []tls.Certificate{cert},
		PreferServerCipherSuites: true,
		ClientCAs:                clientCertsPool,
		ClientAuth:               tls.RequireAndVerifyClientCert,
	}

	listener, err := tls.Listen("tcp", ":8081", tlsCfg)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		log.Println("Waiting for clients")
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("New client: ", conn.RemoteAddr())
		go OnConnection(conn)
	}
}

func main() {
	fmt.Println("A Simple Message board, Support both https[8080] & TLS[8081]")
	Logs = append(Logs, Item{
		IP:         "127.0.0.1",
		Time:       time.Now().String(),
		Text:       "This is a test dashboard demo, support both Https & TLS",
		Connection: "local",
	})

	go TLS()

	http.HandleFunc("/", Hello)
	http.HandleFunc("/Post", Post)
	err := http.ListenAndServeTLS(":8080", "ServerCert.pem", "ServerKey.pem", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
