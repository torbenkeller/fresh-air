package main

import (
	"database/sql"
	"time"

	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"net/http"
)

type registerEventInput struct {
	Sensor string `json:"sensor"`
	Room   string `json:"room"`
	Event  string `json:"event"`
}

func (r registerEventInput) String() string {
	return fmt.Sprintf("{Sensor: %s, Room: %s, Event: %s}", r.Sensor, r.Room, r.Event)
}

func main() {
	connStr := "postgres://postgres:postgres@db:5432/fresh-air?sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}

	r := gin.Default()
	r.POST("/register-event", func(context *gin.Context) {
		registerEvent(context, db)
	})
	r.GET("/shouldOpenWindow/:room", func(context *gin.Context) {
		shouldOpenWindow(context, db)
	})

	r.Run()
}

func registerEvent(c *gin.Context, db *sql.DB) {
	input := registerEventInput{}
	err := c.BindJSON(&input)

	if err != nil {
		c.JSON(http.StatusBadRequest, "Bad Request")
		return
	}

	if input.Sensor == "" || input.Event == "" || input.Room == "" {
		c.JSON(http.StatusBadRequest, "Bad Request")
		return
	}

	_, err = db.Exec(
		"INSERT INTO TB_WINDOW_EVENTS (REGISTERED_AT, SENSOR, EVENT, ROOM) VALUES ($1, $2, $3, $4);",
		time.Now(),
		input.Sensor,
		input.Event,
		input.Room,
	)

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, "Internal Server Error")
		return
	}

	c.JSON(http.StatusOK, "OK")
}

func shouldOpenWindow(c *gin.Context, db *sql.DB) {
	room := c.Param("room")
	res := db.QueryRow("SELECT REGISTERED_AT FROM TB_WINDOW_EVENTS WHERE ROOM = $1 ORDER BY REGISTERED_AT DESC LIMIT 1;", room)
	if res.Err() != nil {
		fmt.Println(res.Err())
		c.JSON(http.StatusInternalServerError, "Internal Server Error")
		return
	}

	var t time.Time
	err := res.Scan(&t)

	if err != nil {
		fmt.Println(res.Err())
		c.JSON(http.StatusInternalServerError, "Internal Server Error")
		return
	}

	now := time.Now()

	if now.Hour() < 7 {
		c.JSON(http.StatusOK, 0)
		return
	}

	if now.After(t.Add(7 * time.Hour)) {
		c.JSON(http.StatusOK, 1)
		return
	}

	c.JSON(http.StatusOK, 0)
	return
}
