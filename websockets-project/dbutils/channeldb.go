package dbutils

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/lib/pq"
	"github.com/segmentio/ksuid"
)

type userChannel struct {
	userID    string
	channelID string
}
type Channel struct {
	Name      string
	Channelid string
	owner     string
	Users     []string
}

//#TODO DO WYJEBANIA
func InitChannelDB() *sql.DB {
	psqlconnection := "postgres://postgres:password@localhost/db?sslmode=disable"

	db, err := sql.Open("postgres", psqlconnection)
	if err != nil {
		fmt.Println("first exception")
		log.Fatal(err)
	}

	err = db.Ping()

	if err != nil {
		fmt.Println("PING EXCEPTION")
		log.Fatal(err)
	}
	return db
}
func QueryChannel(channelName string, db *sql.DB) (Channel, error) {

	c := Channel{}

	rows := db.QueryRow("SELECT CHANNEL_ID, NAME, OWNER, USERS FROM channels WHERE NAME=$1", channelName).Scan(&c.Channelid, &c.Name, &c.owner, &c.Users)
	//#TODO ADD QUERY AND ADD CHANNEL STRUCT HERE
	switch {
	case rows == sql.ErrNoRows:
		return c, sql.ErrNoRows
	default:
		return c, nil
	}
}

func QueryChannelsWhereOwner(username string, db *sql.DB) []Channel {
	c := Channel{}
	channels := make([]Channel, 0)
	res, err := db.Query("Select name, channel_id, owner, users from channels where owner=$1", username)
	if err != nil {
		fmt.Println("could not find chanenls")
	}
	defer res.Close()
	for res.Next() {
		if err := res.Scan(&c.Name, &c.Channelid, &c.owner, pq.Array(&c.Users)); err != nil {
			fmt.Println(err)
		}
		channels = append(channels, c)

	}
	return channels
}
func QueryUsersChannels(userName string, userEmail string, db *sql.DB) []string {
	user := Query(userName, userEmail, true, db)
	channelList := make([]string, 0)
	channel := new(userChannel)
	res, err := db.Query("SELECT USER_ID, CHANNEL_ID FROM USERS_CHANNELS WHERE USER_ID=$1", user.Userid)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Close()
	for res.Next() {
		if err := res.Scan(&channel.userID, &channel.channelID); err != nil {
			fmt.Println(err)
		}
		channelList = append(channelList, channel.channelID)
	}
	return channelList

}

func InsertUsersChannels(userName string, userEmail string, channelName string, db *sql.DB) bool {

	channel, err := QueryChannel(channelName, db)
	if err != nil {
		log.Fatal(err)
	}
	user := Query(userName, userEmail, true, db)
	_, err = db.Exec("INSERT INTO users_channels(user_id, channel_id) VALUES ($1,$2)", user.Userid, channel.Channelid)
	if err != nil {
		log.Fatal(err)
		return false
	}
	return true
}

func checkAvalibility(channelName string, owner string, db *sql.DB) bool {
	c := Channel{}
	rows := db.QueryRow("SELECT CHANNEL_ID, NAME, OWNER, USERS FROM channels WHERE NAME=$1 AND OWNER=$2", channelName, owner).Scan(&c.Channelid, &c.Name, &c.owner, &c.Users)
	if rows == sql.ErrNoRows {
		return true
	}
	return false

}

func CreateChannel(channelName string, owner string, db *sql.DB) bool {
	fmt.Println(owner)
	existingUSer := Doesuserexist(owner, db)
	if avalible := checkAvalibility(channelName, owner, db); avalible && existingUSer {
		id := ksuid.New()
		//u := new(users)
		//u.usernames = []string{owner}
		//v, err := json.Marshal(u)
		c := Channel{}
		c.Users = []string{owner}
		_, err := db.Exec("INSERT INTO channels (CHANNEL_ID, NAME, OWNER, USERS) VALUES($1, $2, $3, $4)", id.String(), channelName, owner, pq.Array(c.Users))
		if err != nil {
			log.Fatal(err)
			return false
		}
		return true

	} else {
		return false
	}
}
func QueryChannelFromID(channelIdList []string, db *sql.DB) []Channel {
	c := Channel{}
	channelList := make([]Channel, 0)
	for i := range channelIdList {
		res := db.QueryRow("SELECT CHANNEL_ID, NAME, OWNER, USERS FROM channels where channel_id=$1", channelIdList[i]).Scan(&c.Channelid, &c.Name, &c.owner, pq.Array(&c.Users))
		if res == sql.ErrNoRows {
			fmt.Println("channel does not exist")
		}
		channelList = append(channelList, c)

	}
	return channelList
}

func Doesuserexist(username string, db *sql.DB) bool {
	rows := db.QueryRow("SELECT username FROM USERS WHERE username=$1", username).Scan(new(string))
	if rows != sql.ErrNoRows {
		return true
	}
	return false
}
func AddUser(usersName string, channel_id string, db *sql.DB) bool {

	channel := QueryChannelFromID([]string{channel_id}, db)[0]
	for _, v := range channel.Users {
		if v == usersName {
			return false
		}
	}
	_, err := db.Exec("UPDATE channels SET users= array_append(users,$1) where channel_id=$2;", usersName, channel_id)
	if err != nil {
		return false
	}
	return true
}

func DeleteUser(usersName string, channel_id string, db *sql.DB) bool {
	channel := QueryChannelFromID([]string{channel_id}, db)[0]
	for _, v := range channel.Users {
		if v == usersName {
			return false
		}
	}
	_, err := db.Exec("UPDATE channels SET users= array_remove(users,$1) where channel_id=$2;", usersName, channel_id)
	if err != nil {
		return false
	}
	return true

}

func InsertChannelsUsers(channelid string, userUSername string, db *sql.DB) bool {
	u := new(Users)
	rows := db.QueryRow("SELECT username,email,password,user_id  FROM USERS WHERE username=$1", userUSername).Scan(&u.Username, &u.Email, &u.Password, &u.Userid)
	if rows == sql.ErrNoRows {
		fmt.Println("user does not exist")
		return false
	}

	_, err := db.Exec("INSERT INTO users_channels(user_id, channel_id) VALUES ($1,$2)", u.Userid, channelid)
	if err != nil {
		fmt.Println("error occured")
		return false
	}
	return true
}
func DeleteChannelUsers(channelid string, userUSername string, db *sql.DB) bool {
	u := new(Users)
	rows := db.QueryRow("SELECT username,email,password,user_id  FROM USERS WHERE username=$1", userUSername).Scan(&u.Username, &u.Email, &u.Password, &u.Userid)
	if rows == sql.ErrNoRows {
		fmt.Println("user does not exist")
		return false
	}
	_, err := db.Exec("DELETE FROM users_channels WHERE user_id=$1 and channel_id=$2", u.Userid, channelid)
	if err != nil {
		fmt.Println("error occured")
		return false
	}
	return true
}
func CheckIfUserIsInChannel(c Channel, username string) bool {

	for _, v := range c.Users {
		if v == username {
			return true
		}
	}
	return false
}
func CheckIfUserIsOwner(username string, channelid string, db *sql.DB) bool {
	channel := QueryChannelFromID([]string{channelid}, db)
	for _, c := range channel {
		if c.owner == username {
			return true
		}

	}
	return false
}
func DeleteChannel(channelid string, username string, db *sql.DB) bool {
	_, err := db.Exec("DELETE from channels where owner=$1 and channel_id=$2;", username, channelid)
	if err != nil {
		log.Print("error occured when deleting ", err)
		return false
	}
	_, err = db.Exec("DELETE from users_channels where channel_id=$1;", channelid)
	if err != nil {
		log.Print("error occured when deleting ", err)
		return false
	}
	return true
}
