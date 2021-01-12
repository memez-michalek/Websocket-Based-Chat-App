package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"sync"
	"text/template"
	"time"

	"github.com/go-session/session"

	"websockets-project/chathandler"
	"websockets-project/dbutils"
	"websockets-project/webtoken"

	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
)

var (
	host       = "localhost"
	port       = 5432
	user       = "postgres"
	password   = "password"
	dbname     = "users"
	sessionKey = []byte("A9qFf/qbWlQTKQ7rg3LT5wHMHn+xPTlxl+zSYnANAf0=")
	ctx        = context.Background()
	wg         = new(sync.WaitGroup)
)
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	templ := template.Must(template.ParseFiles("./templates/main.html"))

	s, err := session.Start(context.Background(), w, r)

	if err != nil {
		fmt.Println(err)
	}
	token, ok := s.Get("webtoken")
	username, ok := s.Get("username")
	email, ok := s.Get("email")
	if ok != true {
		templ.Execute(w, "log in first session['webtoken'] is empty")
	} else if token != nil && token.(string) != "" {

		_, _, err := webtoken.Verify(token.(string))
		if err != nil {
			s.Set("webtoken", "")
			s.Set("username", "")
			s.Set("email", "")
			err := s.Save()
			if err != nil {
				fmt.Println(err)
			}
			http.Redirect(w, r, "/login", 302)

		} else {
			db := dbutils.InitDB()
			allChannelsIDs := dbutils.QueryUsersChannels(username.(string), email.(string), db)
			channels := dbutils.QueryChannelFromID(allChannelsIDs, db)
			u := struct {
				Username      string
				Email         string
				Channelstruct []dbutils.Channel
			}{
				Username:      username.(string),
				Email:         email.(string),
				Channelstruct: channels,
			}

			templ.Execute(w, u)
		}
	} else {
		templ.Execute(w, "token expired not logged in")

	}
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	database := dbutils.InitDB()
	s, err := session.Start(context.Background(), w, r)

	if err != nil {
		log.Fatal(err)

	}

	tkn, e := s.Get("webtoken")
	if e != true {
		http.Redirect(w, r, "/", 302)
		fmt.Println(e)
	}

	if tkn == nil || tkn.(string) == "" {
		if r.Method != http.MethodPost {
			http.ServeFile(w, r, "./templates/register.html")
			//temp.Execute(w, "")
		} else {

			if avalible := dbutils.CheckCredentialsRegister(r.FormValue("username"), r.FormValue("email"), database); avalible {
				dbutils.Insert(r.FormValue("username"), r.FormValue("email"), r.FormValue("password"), database)
				token := webtoken.CreateToken(r.FormValue("username"), r.FormValue("email"))
				s.Set("webtoken", token)
				s.Set("username", r.FormValue("username"))
				s.Set("email", r.FormValue("email"))
				err := s.Save()
				if err != nil {
					fmt.Println(err)
				}
				http.Redirect(w, r, "/", 302)
			} else {
				http.ServeFile(w, r, "./templates/register.html")
			}
		}
	} else {
		http.Redirect(w, r, "/", 302)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	database := dbutils.InitDB()
	s, err := session.Start(context.Background(), w, r)
	if err != nil {
		log.Fatal(err)
	}

	tkn, e := s.Get("webtoken")
	if e != true {

		fmt.Println(e)
	}

	if tkn == nil || tkn.(string) == "" {
		if r.Method != http.MethodPost {
			http.ServeFile(w, r, "./templates/login.html")
		} else {

			_, chkr := dbutils.CheckLoginPassword(r.FormValue("username"), r.FormValue("email"), r.FormValue("password"), database)
			if chkr {
				token := webtoken.CreateToken(r.FormValue("username"), r.FormValue("email"))
				s.Set("webtoken", token)
				s.Set("username", r.FormValue("username"))
				s.Set("email", r.FormValue("email"))
				err := s.Save()

				if err != nil {
					log.Fatal(err)
				}
				http.Redirect(w, r, "/", 302)
			} else {
				http.ServeFile(w, r, "./templates/login.html")
			}
		}
	} else {
		http.Redirect(w, r, "/", 302)
	}

}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	s, err := session.Start(context.Background(), w, r)
	if err != nil {
		log.Fatal(err)
	}
	tkn, e := s.Get("webtoken")

	switch {
	case e != true:
		http.Redirect(w, r, "/login", 302)
	case tkn != nil || tkn.(string) != "":
		fmt.Println("logged out")
		s.Set("webtoken", "")
		s.Set("username", "")
		s.Set("email", "")
		err := s.Save()
		if err != nil {
			fmt.Println(err)
		}
		http.Redirect(w, r, "/", 302)
	default:
		http.Redirect(w, r, "/", 302)
	}

}

func createChannelHandler(w http.ResponseWriter, r *http.Request) {
	s, err := session.Start(context.Background(), w, r)
	if err != nil {
		log.Fatal(err)
	}
	tkn, e := s.Get("webtoken")
	username, e := s.Get("username")
	email, e := s.Get("email")
	switch {
	case e != true:
		fmt.Println("webtoken not found")
		http.Redirect(w, r, "/login", 302)
	case tkn != nil || tkn.(string) != "":
		if r.Method != http.MethodPost {
			http.ServeFile(w, r, "./templates/addchannel.html")
		} else {
			//#TODO CREATE NEW CHANNELS

			db := dbutils.InitChannelDB()
			_, err := dbutils.QueryChannel(r.FormValue("name"), db)
			if err == sql.ErrNoRows {
				_, _, err := webtoken.Verify(tkn.(string))
				fmt.Println(username)
				if err != nil {
					http.Redirect(w, r, "/login", 302)
				}
				created := dbutils.CreateChannel(r.FormValue("name"), username.(string), db)
				inserted := dbutils.InsertUsersChannels(username.(string), email.(string), r.FormValue("name"), db)
				if created && inserted {
					http.Redirect(w, r, "/", 302)
				}

			} else {
				http.ServeFile(w, r, "./templates/addchannel.html")
			}
		}
	default:
		http.Redirect(w, r, "/", 302)
	}

}
func addUserHandler(w http.ResponseWriter, r *http.Request) {
	templ := template.Must(template.ParseFiles("./templates/adduser.html"))
	s, err := session.Start(context.Background(), w, r)
	if err != nil {
		log.Fatal()
	}

	tkn, e := s.Get("webtoken")
	username, e := s.Get("username")
	email, e := s.Get("email")
	if e != true {
		http.Redirect(w, r, "/", 302)
	}
	db := dbutils.InitDB()
	switch {
	case tkn == nil:
		fmt.Println("log in again")
		http.Redirect(w, r, "/login", 302)
	case tkn.(string) == "":
		fmt.Println("log in again")
		http.Redirect(w, r, "/login", 302)
	case username.(string) != "" && email.(string) != "" && tkn.(string) != "":
		_, _, err := webtoken.Verify(tkn.(string))
		if err != nil {
			fmt.Println("token expired")
			http.Redirect(w, r, "/login", 302)
		} else if r.Method != http.MethodPost {
			chans := dbutils.QueryChannelsWhereOwner(username.(string), db)
			if len(chans) == 0 {
				http.Redirect(w, r, "/", 302)
			}
			templ.Execute(w, struct {
				Channels []dbutils.Channel
			}{
				Channels: chans,
			})
		} else {

			exists := dbutils.Doesuserexist(r.FormValue("name"), db)
			if exists != true {
				http.Redirect(w, r, "/adduser", 302)
			}
			added := dbutils.AddUser(r.FormValue("name"), r.FormValue("channelid"), db)
			if added != true {
				http.Redirect(w, r, "/", 302)
				fmt.Println("user has already been added ")
			}
			userschannelsadded := dbutils.InsertChannelsUsers(r.FormValue("channelid"), r.FormValue("name"), db)
			if userschannelsadded != true {
				http.Redirect(w, r, "/", 302)
				fmt.Println("ERROR")
			}
			dir := "/channel/" + r.FormValue("channelid")
			http.Redirect(w, r, dir, 302)
		}

	default:
		http.Redirect(w, r, "/", 302)
	}

}
func deleteUsersFromChannelHandler(w http.ResponseWriter, r *http.Request) {
	templ := template.Must(template.ParseFiles("./templates/deleteuser.html"))
	s, err := session.Start(context.Background(), w, r)
	if err != nil {
		log.Fatal()
	}

	tkn, e := s.Get("webtoken")
	username, e := s.Get("username")
	email, e := s.Get("email")
	if e != true {
		http.Redirect(w, r, "/", 302)
	}
	db := dbutils.InitDB()
	switch {

	case tkn == nil:
		fmt.Println("user is not logged in")
		http.Redirect(w, r, "/login", 302)

	case tkn.(string) == "":

		fmt.Println("user is not logged in")
		http.Redirect(w, r, "/login", 302)

	case username.(string) != "" && email.(string) != "" && tkn.(string) != "":
		_, _, err := webtoken.Verify(tkn.(string))
		if err != nil {
			fmt.Println("token expired")
			http.Redirect(w, r, "/login", 302)
		} else if r.Method != http.MethodPost {
			channels := dbutils.QueryChannelsWhereOwner(username.(string), db)
			templ.Execute(w, struct {
				Channels []dbutils.Channel
			}{
				Channels: channels,
			})
		} else {
			exists := dbutils.Doesuserexist(r.FormValue("name"), db)
			if exists != true {
				http.Redirect(w, r, "/deleteuser", 302)
			}
			deleted := dbutils.DeleteUser(r.FormValue("name"), r.FormValue("channelid"), db)
			if deleted != true {
				http.Redirect(w, r, "/", 302)
			}
			deleteduserchannel := dbutils.DeleteChannelUsers(r.FormValue("channelid"), r.FormValue("name"), db)
			if deleteduserchannel != true {
				http.Redirect(w, r, "/", 302)
				fmt.Println("ERROR!")
			}
			dir := "/channel/" + r.FormValue("channelid")
			http.Redirect(w, r, dir, 302)
		}

	default:
		http.Redirect(w, r, "/", 302)
	}
}
func showChannel(w http.ResponseWriter, r *http.Request) {
	db := dbutils.InitDB()
	client := dbutils.InitRedisConn()
	channelID := r.URL.Path[len("/channel/"):]
	channel := dbutils.QueryChannelFromID([]string{channelID}, db)[0]
	templ := template.Must(template.ParseFiles("./templates/channel.html"))
	s, err := session.Start(context.Background(), w, r)

	if err != nil {
		http.Redirect(w, r, "/login", 302)
	}
	tkn, e := s.Get("webtoken")
	username, e := s.Get("username")

	switch {
	case tkn == nil:
		fmt.Println("webtoken is empty")
		http.Redirect(w, r, "/login", 302)
	case !e:
		fmt.Println("webtoken is empty")
		http.Redirect(w, r, "/login", 302)
	case tkn.(string) == "":
		http.Redirect(w, r, "/login", 302)
		fmt.Println("webtoken is empty")
	case dbutils.CheckIfUserIsInChannel(channel, username.(string)) == false:
		http.Redirect(w, r, "/", 302)
	}
	_, _, err = webtoken.Verify(tkn.(string))
	if err != nil {
		http.Redirect(w, r, "/login", 302)
		fmt.Println("token expired")
	}

	messages, err := client.LRange(ctx, channel.Channelid, 0, 100).Result()
	if err != nil {
		log.Fatal(err)
	}
	templ.Execute(w, struct {
		Messages []string
		Channel  dbutils.Channel
	}{
		Messages: messages,
		Channel:  channel,
	})

}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	s, err := session.Start(ctx, w, r)
	if err != nil {
		log.Fatal(err)
	}
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
	}
	channelID := r.URL.Path[len("/chat/"):]
	username, e := s.Get("username")
	if e != true {
		log.Fatal(e)
	}
	client := chathandler.MakeClient(username.(string), conn)
	//channel, err := chathandler.GetChannelFromID(channelID)
	channel, ok := chathandler.CHANNELLIST[channelID]
	if !ok {
		log.Print(err)
		channel = chathandler.MakeChannel(channelID)
		channel.Websockets = append(channel.Websockets, client)
		chathandler.CHANNELLIST[channelID] = channel
	} else {
		channel.Websockets = append(channel.Websockets, client)

	}
	conn.SetReadDeadline(time.Now().Add(90 * time.Second))
	conn.SetPongHandler(func(string) error { conn.SetReadDeadline(time.Now().Add(90 * time.Second)); return nil })

	go channel.BROADCASTMessages()
	wg.Add(2)
	go client.MessageWriter(channelID)
	go client.RecvLocalMessages(channelID)
	wg.Done()
}
func quitChannelHandler(w http.ResponseWriter, r *http.Request) {
	db := dbutils.InitDB()
	templ := template.Must(template.ParseFiles("./templates/quitchannel.html"))
	s, err := session.Start(ctx, w, r)
	if err != nil {
		log.Print("session error")
	}
	username, ok := s.Get("username")
	email, ok := s.Get("email")
	token, ok := s.Get("webtoken")
	switch {
	case !ok:
		log.Print("session error username/webtoken")
		http.Redirect(w, r, "/login", 302)
	case token == nil || username == nil:
		log.Print("token is nil")
		http.Redirect(w, r, "/login", 302)
	case token.(string) == "" || username.(string) == "":
		log.Print("token is nil")
		http.Redirect(w, r, "/login", 302)
	}
	_, _, err = webtoken.Verify(token.(string))
	if err != nil {
		log.Print("webtoken expired")
		http.Redirect(w, r, "/login", 302)
	}

	TemplateChannels := make([]dbutils.Channel, 0)
	channelsList := dbutils.QueryUsersChannels(username.(string), email.(string), db)
	for _, v := range dbutils.QueryChannelFromID(channelsList, db) {
		TemplateChannels = append(TemplateChannels, v)
	}
	if r.Method != http.MethodPost {
		templ.Execute(w, struct {
			Channels []dbutils.Channel
		}{
			Channels: TemplateChannels,
		})
	} else {
		id := r.FormValue("id")
		isowner := dbutils.CheckIfUserIsOwner(username.(string), id, db)
		if !isowner {
			deleted := dbutils.DeleteUser(username.(string), id, db)
			if !deleted {
				log.Print("an error occured while deleting the user")
				http.Redirect(w, r, "/", 302)
			}
			deleted = dbutils.DeleteChannelUsers(id, username.(string), db)
			if !deleted {
				log.Print("an error occured while deleting the user")
				http.Redirect(w, r, "/", 302)
			}
			http.Redirect(w, r, "/", 302)

		} else {
			delete := dbutils.DeleteChannel(id, username.(string), db)
			if !delete {
				log.Print("error")
			}
			http.Redirect(w, r, "/", 302)
		}
	}

}
func deleteChannelHandler(w http.ResponseWriter, r *http.Request) {
	db := dbutils.InitDB()
	templ := template.Must(template.ParseFiles("./templates/quitchannel.html"))
	s, err := session.Start(ctx, w, r)
	if err != nil {
		log.Print("session error")
	}
	username, ok := s.Get("username")
	token, ok := s.Get("webtoken")
	switch {
	case !ok:
		log.Print("session error username/webtoken")
		http.Redirect(w, r, "/login", 302)
	case token == nil || username == nil:
		log.Print("token is nil")
		http.Redirect(w, r, "/login", 302)
	case token.(string) == "" || username.(string) == "":
		log.Print("token is nil")
		http.Redirect(w, r, "/login", 302)
	}
	_, _, err = webtoken.Verify(token.(string))
	if err != nil {
		log.Print("webtoken expired")
		http.Redirect(w, r, "/login", 302)
	}
	channels := dbutils.QueryChannelsWhereOwner(username.(string), db)
	if r.Method != http.MethodPost {
		templ.Execute(w, struct {
			Channels []dbutils.Channel
		}{
			Channels: channels,
		})
	} else {
		id := r.FormValue("id")
		deleted := dbutils.DeleteChannel(id, username.(string), db)
		if !deleted {
			log.Print("error occured when deleting channel")
			http.Redirect(w, r, "/", 302)
		}
		http.Redirect(w, r, "/", 302)
	}

}
func main() {

	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/login/", loginHandler)
	http.HandleFunc("/register/", registerHandler)
	http.HandleFunc("/logout/", logoutHandler)
	http.HandleFunc("/createchannel/", createChannelHandler)
	http.HandleFunc("/channel/", showChannel)
	http.HandleFunc("/adduser/", addUserHandler)
	http.HandleFunc("/deleteuser/", deleteUsersFromChannelHandler)
	http.HandleFunc("/chat/", chatHandler)
	http.HandleFunc("/quit/", quitChannelHandler)
	http.HandleFunc("/deletechannel/", deleteChannelHandler)

	http.ListenAndServe(":8080", nil)
}
