package dbutils

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/segmentio/ksuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "password"
	dbname   = "users"
)

type Users struct {
	Username string
	Password []byte
	Email    string
	Userid   string
}

func InitDB() *sql.DB {
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
func Query(username string, email string, queryTypeLogin bool, database *sql.DB) Users {
	var (
		tempMail     string
		tempUsername string
		tempPasswd   []byte
		user_id      string
	)

	res := database.QueryRow(`select username, email, password, user_id from users where username=$1 and email=$2;`, username, email).Scan(&tempMail, &tempUsername, &tempPasswd, &user_id)
	u := Users{}

	switch {
	case res == sql.ErrNoRows && queryTypeLogin == false:
		u := Users{
			Username: "",
			Email:    "",
			Password: []byte(""),
			Userid:   "",
		}
		return u

	case res != sql.ErrNoRows && queryTypeLogin == true:
		u := Users{
			Username: tempUsername,
			Email:    tempMail,
			Password: tempPasswd,
			Userid:   user_id,
		}
		return u
	case res == sql.ErrNoRows && queryTypeLogin == true:
		u := Users{
			Username: "",
			Email:    "",
			Password: []byte(""),
			Userid:   "",
		}
		return u
	case res != nil:
		log.Fatal(res)

	}
	return u

}

func Insert(username string, email string, password string, db *sql.DB) {
	//hash := sha512.Sum512([]byte(password))
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		log.Fatal(err)
	}
	user_id := ksuid.New()
	_, err = db.Exec(`INSERT INTO users(username, email, password, user_id) VALUES ($1, $2, $3, $4)`, username, email, hash, user_id.String())
	if err != nil {
		log.Fatal(err)

		defer db.Close()
	}
}

func CheckLoginPassword(username string, email string, password string, database *sql.DB) (Users, bool) {
	template := Users{
		Email:    "",
		Password: []byte(""),
		Username: "",
		Userid:   "",
	}

	dbUser := Query(username, email, true, database)
	if dbUser.Email == "" && dbUser.Username == "" {
		return template, false
	}

	match := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(password))
	defer database.Close()

	if dbUser.Email == "" || dbUser.Username == "" || match != nil {
		fmt.Println("wrong")
		return template, false
	} else {
		fmt.Println("correct")
		return dbUser, true
	}
}
func CheckCredentialsRegister(username string, email string, db *sql.DB) bool {
	u := Query(username, email, false, db)
	if u.Email == "" || u.Username == "" {
		return true
	} else {
		return false
	}

}
