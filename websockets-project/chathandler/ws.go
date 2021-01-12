package chathandler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"websockets-project/dbutils"

	"github.com/gorilla/websocket"
)

var ctx = context.Background()

type Client struct {
	Username     string
	Sock         *websocket.Conn
	Messageschan chan []byte
}

type Message struct {
	sender  Client
	time    int64
	content string
}
type Channel struct {
	Id         string
	Websockets []Client
	Messages   chan []byte
}

var CHANNELLIST = make(map[string]*Channel, 0)

func (client *Client) unregisterUser(channelID string) {
	table := make([]Client, 0)
	channel := CHANNELLIST[channelID]
	for i, v := range channel.Websockets {
		if v.Username == client.Username {
			if len(CHANNELLIST[channelID].Websockets) < 2 {
				CHANNELLIST[channelID].Websockets = table
			} else {

				table = append(table, CHANNELLIST[channelID].Websockets[:i]...)
				table = append(table, CHANNELLIST[channelID].Websockets[i+1:]...)
				CHANNELLIST[channelID].Websockets = table
			}
		}
	}

}

func MakeClient(username string, conn *websocket.Conn) Client {
	c := new(Client)
	c.Sock = conn
	c.Username = username
	c.Messageschan = make(chan []byte)
	return *c
}
func MakeChannel(channelid string) *Channel {
	c := new(Channel)
	c.Id = channelid
	c.Messages = make(chan []byte)
	c.Websockets = []Client{}
	return c
}

func findChannelsClients(channelid string) ([]Client, error) {
	channel, err := CHANNELLIST[channelid]
	if !err {
		log.Print("error occured")
		return nil, errors.New("no channel with such id")
	}
	return channel.Websockets, nil
}

func (client *Client) RecvLocalMessages(channelid string) {

	channel, err := CHANNELLIST[channelid]
	if !err {
		log.Print("channel does not exist")
	}
	for {
		_, msg, err := client.Sock.ReadMessage()
		if err != nil {
			fmt.Println(err)
			client.unregisterUser(channelid)
			return
		}
		log.Print(string(msg), " <-recv func ->", client.Username)
		channel.Messages <- msg
	}

}

func (channel *Channel) BROADCASTMessages() {

	redis := dbutils.InitRedisConn()
	for {

		select {
		case msg, ok := <-channel.Messages:
			if ok != true {
				log.Print(ok)
			}
			redis.LPush(ctx, channel.Id, string(msg)).Result()
			for _, client := range channel.Websockets {
				fmt.Println("MSG BROADCAST ->", string(msg))
				client.Messageschan <- msg

			}

		}

	}

}
func (c *Client) MessageWriter(channelid string) {
	ticker := time.NewTicker(time.Second * 60)
	defer func() {
		ticker.Stop()
		c.Sock.Close()
	}()

	for {
		select {
		case msg, ok := <-c.Messageschan:
			if !ok {
				log.Print(ok)

			}
			fmt.Println("MESSAGEWRITER->", string(msg), " ", c.Username)
			err := c.Sock.WriteMessage(1, msg)
			if err != nil {
				log.Print(err)
				c.unregisterUser(channelid)
				return
			}
		case <-ticker.C:
			c.Sock.SetWriteDeadline(time.Now().Add(time.Second * 60))
			if err := c.Sock.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}

}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
