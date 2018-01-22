package main

import (
  "github.com/heshoots/mmr/schema/users"
  "github.com/heshoots/mmr/schema/match"
  "database/sql"
  _ "github.com/lib/pq"
  "log"
  "github.com/aspic/go-challonge"
  "os"
  "text/template"
  "errors"
  "fmt"
  "bytes"
  "strconv"
  "gopkg.in/echo.v3"
)

func errCheck(err error) {
  if err != nil {
    log.Fatal(err)
  }
}

func handler(c echo.Context) error {
    connStr := "user=postgres sslmode=disable"
    db, err := sql.Open("postgres", connStr)
    defer db.Close()
    errCheck(err)
    c.JSONBlob(200, []byte("{"))
    games, err := db.Query("SELECT game_id, name FROM games");
    for games.Next() {
	    var game_id int
	    var game_name string
	    games.Scan(&game_id, &game_name)
	    c.JSONBlob(200, []byte(game_name))
	    c.JSONBlob(200, []byte(":"))
	    rows, _ := db.Query("SELECT display_name, elo FROM users, elo WHERE elo.user_id = users.user_id AND elo.game_id=$1 ORDER BY elo", game_id)
	    c.JSONBlob(200, []byte(",["))
	    for rows.Next() {
	      var player users.User
	      rows.Scan(&player.Name, &player.Elo)
	      tmpl, err := template.New("user").Parse("{name: {{.Name}}, elo: {{.Elo}}}")
	      if err != nil {
		fmt.Println(err)
	      }
	      var result bytes.Buffer
	      tmpl.Execute(&result, player)
	      c.JSONBlob(200, []byte(result.String()))
	      //w.Write([]byte("namele server.\n"))
	    }
	    c.JSONBlob(200, []byte("]"))
    }
    c.JSONBlob(200, []byte("}"))
    return errors.New("")
}

func deleteSchema(db *sql.DB) {
  db.Query("DROP TABLE users CASCADE");
  db.Query("DROP TABLE games CASCADE");
  db.Query("DROP TABLE elo");
  db.Query("DROP TABLE match");
}

func createSchema(db *sql.DB) {
  _, err := db.Query("CREATE TABLE games (game_id SERIAL, name VARCHAR(100), PRIMARY KEY (game_id))")
  _, err = db.Query("CREATE TABLE users (user_id SERIAL, display_name VARCHAR(50), PRIMARY KEY (user_id))");
  errCheck(err)
  _, err = db.Query("CREATE TABLE elo (elo_id SERIAL, game_id INTEGER REFERENCES games(game_id), user_id INTEGER REFERENCES users(user_id), elo float, PRIMARY KEY (elo_id))");
  errCheck(err)
  _, err = db.Query("CREATE TABLE match (match_id SERIAL, game_id INTEGER REFERENCES games(game_id), date TIMESTAMP, player1_id INTEGER REFERENCES users(user_id), player2_id INTEGER REFERENCES users(user_id), player1_score INTEGER, player2_score INTEGER)")
  errCheck(err)
}

func main() {
  connStr := "user=postgres sslmode=disable"
  db, err := sql.Open("postgres", connStr)
  defer db.Close()
  errCheck(err)

  argsWithoutProg := os.Args[1:]
  arg := argsWithoutProg[0]
  if arg == "clear" {
    deleteSchema(db)
    createSchema(db)
  }
  if arg == "fill" {
    tourney := argsWithoutProg[1]

    client := challonge.New("Quora", os.Getenv("CHALLONGE_API_KEY"))
    t, err := client.NewTournamentRequest(tourney).WithMatches().WithParticipants().Get()
    errCheck(err)
    fmt.Println(t.GameName)
    _, err = db.Query("INSERT INTO games (name) VALUES ($1) ON CONFLICT DO NOTHING;", t.GameName)
    errCheck(err)
    var gameId int
    err = db.QueryRow("SELECT game_id FROM games WHERE name=$1", t.GameName).Scan(&gameId)
    errCheck(err)
    for i := 0; i < len(t.Matches); i++ {
      match.ReportMatch(db, gameId, t.Matches[i])
    }
  }
  if arg == "combine" {
    keep := argsWithoutProg[1]
    remove := argsWithoutProg[2]
    gameId, _ := strconv.ParseInt(argsWithoutProg[3], 10, 64)
    users.CombineUser(db, keep, remove, int(gameId))
  }
  if arg == "calc" {

    match.CalculateElo(db)
  }
  if arg == "server" {
    e := echo.New()
    e.GET("/list", handler)
    e.Start(":8000")
  }
}
