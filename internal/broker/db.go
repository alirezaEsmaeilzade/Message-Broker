package broker

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"therealbroker/pkg/broker"
	"time"
)

type Psql struct{
	client *sql.DB
}
type DB interface {
	AddMessage(subject string, message *broker.Message) (int, error)
	FetchMessage(ID int) (*broker.Message, error)
	DeleteMessage(ID int) error
}

func NewPsql() DB{
	connString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		"127.0.0.1", "5432", "postgres", "myPassword", "brokerdb")
	db, err := sql.Open("postgres", connString)
	//defer db.Close()
	if err != nil{
		log.Fatal("error can not connect to database")
	}
	_, er := db.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
				ID 		 SERIAL,
				Subject  varchar(255),
				Body	 varchar,
				Timeout  int4,
				PRIMARY KEY (ID, Subject)
		);
	`)
	db.SetMaxOpenConns(50) /// new added
	//db.SetMaxIdleConns(10)
	if er != nil{
		log.Fatal(er)
	}
	if err != nil{
		log.Fatal("error")
	}
	//log.Println("created")
	return &Psql{client: db}
}

func (psql *Psql) AddMessage(subject string, message *broker.Message) (int, error){
	//fmt.Println("addmessage db")
	var msgID int
	rows, err := psql.client.Query(`
	INSERT INTO messages (Subject, Body, Timeout)
	VALUES ($1, $2, $3)
	RETURNING ID;
	`, subject,message.Body, int32(message.Expiration.Seconds()))
	if err != nil{
		return -1, err
	}
	defer rows.Close()
	rows.Next()
	err = rows.Scan(&msgID)
	if err != nil{
		log.Println(err)
	}
	return msgID, err
}
func (psql *Psql) FetchMessage(ID int) (*broker.Message, error) {
	//fmt.Println("Fetch db")
	var timeout int
	var body string
	rows, err := psql.client.Query(`
	SELECT Body, Timeout FROM messages WHERE ID = $1;
	`, ID)
	if err != nil{
		return &broker.Message{Body: ""}, err
	}
	defer rows.Close()
	rows.Next()
	err = rows.Scan(&body, &timeout)
	if err != nil{
		return &broker.Message{Body: ""}, err
	}
	return &broker.Message{Body: body, Expiration: time.Duration(timeout) * time.Second}, nil
}
func (psql *Psql) DeleteMessage(ID int) error{
	//fmt.Println("delete db")
	_, err := psql.client.Exec(`
		DELETE FROM messages
		WHERE ID = $1;
		`, ID)
	return err
}

