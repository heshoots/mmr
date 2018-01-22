package match

import (
  "github.com/heshoots/mmr/mmr"
  "github.com/heshoots/mmr/schema/users"
  "github.com/aspic/go-challonge"
  "database/sql"
  "strings"
  "strconv"
)

func ReportMatch(db *sql.DB, gameId int, match *challonge.Match) error {
  scores := strings.Split(match.Scores, "-")
  p1Score, err := strconv.ParseFloat(scores[0], 10)
  p2Score, err := strconv.ParseFloat(scores[1], 10)
  player1 := users.GetOrCreate(db, match.PlayerOne.Name, gameId)
  player2 := users.GetOrCreate(db, match.PlayerTwo.Name, gameId)
  on, err := db.Query("INSERT INTO match (date, game_id, player1_id, player2_id, player1_score, player2_score) VALUES ($1, $2, $3, $4, $5, $6)", match.UpdatedAt, gameId, player1.Uid, player2.Uid, p1Score, p2Score)
  defer on.Close()
  return err
}

func CalculateElo(db *sql.DB) {
  // Set all elo's to 1200
  db.Query("update elo set elo=1200")
  games, _ := db.Query("SELECT game_id from games")
  for games.Next() {
    var game_id int
    games.Scan(&game_id)
    rows, _ := db.Query("SELECT date, player1_id, player2_id, player1_score, player2_score from match  WHERE game_id = $1 ORDER BY date", game_id)
    var date string
    var p1_id int
    var p2_id int
    var p1_score float64
    var p2_score float64
    for rows.Next() {
      rows.Scan(&date, &p1_id, &p2_id, &p1_score, &p2_score)
      eloRowp1 := db.QueryRow("select elo from elo WHERE user_id = $1", p1_id)
      var p1_elo float64
      eloRowp1.Scan(&p1_elo)
      eloRowp2 := db.QueryRow("select elo from elo WHERE user_id = $1", p2_id)
      var p2_elo float64
      eloRowp2.Scan(&p2_elo)
      new_p1_elo := mmr.NewRating(p1_elo, p2_elo, p1_score, p2_score)
      new_p2_elo := mmr.NewRating(p2_elo, p1_elo, p2_score, p1_score)
      db.Query("UPDATE elo set elo=$1 WHERE user_id=$2", new_p1_elo, p1_id)
      db.Query("UPDATE elo set elo=$1 WHERE user_id=$2", new_p2_elo, p2_id)
    }
  }
}
