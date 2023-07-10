package server

import (
	"golang.org/x/net/websocket"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

type Connect struct {
	id        int
	conn      net.Conn
	lastVisit int64
}

var (
	connsIndex = 0
	conns      = []Connect{}
)

func wsHandler(conn *websocket.Conn) {
	times := 0
	for {
		data := make([]byte, 1024)
		if _, err := conn.Read(data); err != nil {
			break
		}
		if data == nil || len(data) == 0 {
			continue
		}
		if data[0] == 2 { //dial :
			lenOfData := int(data[1])
			url := string(data[2 : 2+lenOfData])
			d, err := net.Dial(url[0:strings.Index(url, "://")], url[strings.Index(url, "://")+3:])
			if err != nil {
				break
			}
			data = []byte{0}
			if err == nil {
				connsIndex++
				ci := connsIndex
				conns = append(conns, Connect{id: ci, conn: d, lastVisit: time.Now().UnixNano()})
				data = []byte{1, byte(ci)}
			}
			if _, err = conn.Write(data); err != nil {
				break
			}
			continue
		} else if data[0] == 4 { //ResolveIPAddr
			ci := int(data[1])
			name := string(data[2 : 2+ci])
			addr, err := net.ResolveIPAddr("ip", name)
			if err != nil {
				_, err = conn.Write([]byte{0})
				if err != nil {
					log.Println(err.Error())
					break
				}
			} else {
				b := []byte{}
				b = append(b, byte(1))
				to4 := addr.IP.To4()
				b = append(b, to4...)
				if _, err = conn.Write(data); err != nil {
					log.Println(err.Error())
					break
				}
			}
			continue
		} else if data[0] == 3 { //tunneling
			ci := int(data[1])
			var c net.Conn
			for i, connect := range conns {
				if connect.id == ci {
					c = connect.conn
					cs := conns[0:i]
					cs = append(cs, conns[i+1:]...)
					conns = cs
					break
				}
			}
			if c != nil {
				log.Println("tunneling ...")
				_, _ = conn.Write([]byte{1})
			} else {
				log.Println("tunneling not found")
				_, _ = conn.Write([]byte{0})
				continue
			}
			//go
			go func() { io.Copy(c, conn) }()
			io.Copy(conn, c)
			continue

		} else if data[0] == 33 {
			ci := int(data[1])
			name := string(data[2 : 2+ci])
			c, err := net.Dial(name[0:strings.Index(name, "://")], name[strings.Index(name, "://")+3:])
			if err != nil {
				_, _ = conn.Write([]byte{0})
				break
			}
			_, _ = conn.Write([]byte{1})
			go func() { io.Copy(c, conn) }()
			io.Copy(conn, c)
			break

		} else if data[0] == 1 { //ping
			_, _ = conn.Write([]byte{1})
			continue
		}
		times++
	}
}
func Start(l string) error {
	http.Handle("/ws", websocket.Handler(wsHandler))
	log.Println("start websocket server on :", l)
	return http.ListenAndServe(l, nil)
}
