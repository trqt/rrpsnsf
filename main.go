package main

import (
	"database/sql"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	_ "github.com/mattn/go-sqlite3"
)

type Riddle struct {
	MajorTip string
	MinorTip string
	ImgUrl   string
	Answer   string
}

func loadRiddle(db *sql.DB, id int) (Riddle, error) {
	stmt, err := db.Prepare("select major_tip, minor_tip, img_url from riddles where id = ?")
	defer stmt.Close()

	var riddle Riddle
	err = stmt.QueryRow(id).Scan(&riddle.MajorTip, &riddle.MinorTip, &riddle.ImgUrl)

	return riddle, err
}

func cleanGuess(guess string) string {
	return strings.ToLower(strings.TrimSpace(guess))
}

func main() {
	counter := 1 // TODO: proper auth
	db, err := sql.Open("sqlite3", "./riddles.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	engine := html.New("./views", ".html")
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Get("/", func(c *fiber.Ctx) error {
		riddle, err := loadRiddle(db, counter)
		if err != nil {
			log.Println(err)
			return fiber.ErrInternalServerError
		}
		return c.Render("index", fiber.Map{
			"MainTip":  riddle.MajorTip,
			"MinorTip": riddle.MinorTip,
			"ImgUrl":   riddle.ImgUrl,
		})
	})

	app.Post("/riddle", func(c *fiber.Ctx) error {
		guess := c.FormValue("guess")

		stmt, err := db.Prepare("select answer from riddles where id = ?")
		if err != nil {
			log.Println(err)
			return fiber.ErrInternalServerError
		}
		var correct_guess string
		err = stmt.QueryRow(counter).Scan(&correct_guess)
		if err != nil {
			log.Println(err)
			return fiber.ErrInternalServerError
		}

		if cleanGuess(guess) != cleanGuess(correct_guess) {
			return fiber.ErrUnauthorized
		}
		counter++ // FIX: racing condition do it proper

		riddle, err := loadRiddle(db, counter)
		if err != nil {
			log.Println(err)
			return fiber.ErrInternalServerError
		}
		return c.Render("riddle", fiber.Map{
			"MainTip":  riddle.MajorTip,
			"MinorTip": riddle.MinorTip,
			"ImgUrl":   riddle.ImgUrl,
		})
	})

	log.Fatal(app.Listen(":3000"))
}
