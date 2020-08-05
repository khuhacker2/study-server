package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
)

var chatRooms = map[uint64]*ChatRoom{}
var hubMutex = new(sync.Mutex)
var completeCh = make(chan int)

type ChatRoom struct {
	group    uint64
	sessions []*Session

	broadcastCh chan []byte
	enter       chan *Session
	exit        chan *Session
}

func (room *ChatRoom) pump() {
	for {
		select {
		case message := <-room.broadcastCh:
			for _, session := range room.sessions {
				session.write(message)
			}

		case session := <-room.enter:
			room.sessions = append(room.sessions, session)
			completeCh <- len(room.sessions)
		case session := <-room.exit:
			for i, sess := range room.sessions {
				if sess == session {
					room.sessions = append(room.sessions[:i], room.sessions[i+1:]...)
					break
				}
			}

			completeCh <- len(room.sessions)
		}
	}
}

func (room *ChatRoom) broadcast(message []byte) {
	room.broadcastCh <- message
}

type Session struct {
	user uint64
	conn *websocket.Conn
	room *ChatRoom

	send chan []byte
}

func (session *Session) readPump() {
	defer func() {
		session.leave()
		session.conn.Close()
	}()

	session.conn.SetReadDeadline(time.Now().Add(20 * time.Second))
	session.conn.SetPongHandler(func(string) error { session.conn.SetReadDeadline(time.Now().Add(20 * time.Second)); return nil })
	for {
		_, message, err := session.conn.ReadMessage()
		if err != nil {
			break
		}

		fmt.Println(string(message))

		root := gjson.ParseBytes(message)
		cmd := root.Get("cmd").String()
		switch cmd {
		case "sign":
			session.sign(root.Get("data").String())
		case "join":
			session.join(root.Get("data").Uint())
			session.write(message)
		case "chat":
			session.chat(root.Get("data").String())
		}
	}
}

func (session *Session) writePump() {
	defer func() {
		session.conn.Close()
	}()

	ticker := time.NewTicker(10 * time.Second)

	fmt.Println("writeStart")
	for {
		select {
		case message := <-session.send:
			session.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

			w, err := session.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			session.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := session.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (session *Session) write(message []byte) {
	fmt.Println("si")
	session.send <- message
	fmt.Println("bal")
}

func (session *Session) sign(token string) {
	no, ok := parseToken(token)
	if !ok {
		session.conn.Close()
		return
	}

	session.user = no
	data, _ := json.Marshal(struct {
		Cmd  string `json:"cmd"`
		Data interface{}
	}{
		Cmd:  "sign",
		Data: no,
	})
	session.write(data)
	fmt.Println(string(data))
}

func (session *Session) join(group uint64) {
	if session.user == 0 || session.room != nil {
		return
	}

	joined := 0
	database.NewSession(nil).Select("1").From("study_members").Where("user=?", session.user).Load(&joined)
	if joined == 0 {
		return
	}

	hubMutex.Lock()
	defer hubMutex.Unlock()

	if room, ok := chatRooms[group]; ok {
		session.room = room
		room.enter <- session
		<-completeCh
	} else {
		exist := 0
		database.NewSession(nil).Select("1").From("studygroups").Where("no=?", group).Load(&exist)

		if exist == 1 {
			room = &ChatRoom{group: group, sessions: []*Session{session}, broadcastCh: make(chan []byte), enter: make(chan *Session), exit: make(chan *Session)}
			chatRooms[group] = room
			session.room = room
			go room.pump()
		}
	}
}

func (session *Session) leave() {
	if session.room == nil {
		return
	}

	hubMutex.Lock()
	defer hubMutex.Unlock()

	session.room.exit <- session
	members := <-completeCh
	if members == 0 {
		if _, ok := chatRooms[session.room.group]; ok {
			delete(chatRooms, session.room.group)
		}
	}

	session.room = nil
	session.conn.Close()
}

func (session *Session) chat(message string) {
	if session.room == nil {
		return
	}

	data, _ := json.Marshal(struct {
		Cmd  string      `json:"cmd"`
		Data interface{} `json:"data"`
	}{
		Cmd: "chat",
		Data: struct {
			User    uint64 `json:"user"`
			Message string `json:"message"`
		}{
			User:    session.user,
			Message: message,
		},
	})

	session.room.broadcast(data)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func GetWS(w rest.ResponseWriter, r *rest.Request) {
	conn, _ := upgrader.Upgrade(w.(http.ResponseWriter), r.Request, nil)
	session := Session{conn: conn, send: make(chan []byte)}

	go session.writePump()
	go session.readPump()
}
