package users

import (
  "database/sql"
  "log"
  "fmt"
)

type User struct {
  Uid int
  Name string
  Elo float64
}

func DeleteUser(db *sql.DB, user User) {
    db.Query("delete from elo where user_id=$1", user.Uid)
    db.Query("delete from users where user_id=$1", user.Uid)
}

func CombineUser(db *sql.DB, keep string, remove string, gameId int) {
    playerKeep, err := GetUser(db, keep, gameId)
    if err != nil {
      fmt.Println(err)
    }
    playerRemove, err := GetUser(db, remove, gameId)
    if err != nil {
      fmt.Println(err)
    }
    _, err = db.Query("update match set player1_id=$1 where player1_id=$2", playerKeep.Uid, playerRemove.Uid)
    if err != nil {
      fmt.Println(err)
    }
    db.Query("update match set player2_id=$1 where player2_id=$2", playerKeep.Uid, playerRemove.Uid)
    DeleteUser(db, playerRemove)
}

func GetOrCreate(db *sql.DB, name string, gameId int) User {
  var user User
  user, err := GetUser(db, name, gameId)
  if err != nil {
    fmt.Println(err)
    creationErr := CreateUser(db, name, gameId)
    if creationErr != nil {
      log.Fatal(err)
    }
    user, err = GetUser(db, name, gameId)
    if err != nil {
      log.Fatal(err)
    }
  }
  return user
}

func errCheck(err error) {
  if err != nil {
    log.Fatal(err)
  }
}

func CreateUser(db *sql.DB, name string, gameId int) error {
  _, err := db.Query("INSERT INTO users (display_name) values ($1)", name)
  row := db.QueryRow("SELECT user_id FROM users WHERE display_name=$1", name)
  var id int
  err = row.Scan(&id)
  db.Query("INSERT INTO elo (user_id, elo, game_id) VALUES ($1, $2, $3)", id, 1200.0, gameId)
  return err
}

func GetUser(db *sql.DB, name string, gameId int) (User, error) {
  row := db.QueryRow("SELECT users.user_id, display_name, elo  FROM users, elo WHERE users.display_name=$1 AND elo.game_id=$2 AND users.user_id = elo.user_id", name, gameId)
  var user User
  err := row.Scan(&user.Uid, &user.Name, &user.Elo)
  return user, err
}

func UpdateUser(db *sql.DB, user User, gameId int) error {
  _, err:= db.Query("update elo set elo=$1 WHERE user_id=$2 AND game_id=$3", user.Elo, user.Uid, gameId)
  return err
}
