# Websocket-Based-Chat-App

An app which allows you to communicate with others in real time.  

Communication between users is based around websockets. Messages are stored via redis and POSTGRES is used for main database

# HOW TO START THE APP?

First of all you need to start redis and postgres db.

docker-compose up --build

Then you need to change directory to websockets-project and run main.go:

cd websockets-project

go run main.go
