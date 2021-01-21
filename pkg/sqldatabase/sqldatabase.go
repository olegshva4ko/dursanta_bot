package sqldatabase

import (
	"database/sql"
	"dursanta/pkg/models"
	"errors"
	"fmt"
	"math/rand"
	"time"

	_ "github.com/go-sql-driver/mysql" //mysql driver
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

//Sqldatabase allows you to work with sql
type Sqldatabase struct {
	Bot      *tgbotapi.BotAPI
	SQLtoken string
}

//CheckDB checks if connection can be established
func (mydb *Sqldatabase) CheckDB() error {
	db, err := sql.Open("mysql", mydb.SQLtoken)
	if err != nil {
		fmt.Println("line 22", err.Error())
		return err
	}
	if err := db.Ping(); err != nil {
		return err
	}
	db.Close()
	return nil
}

//AddUser this function adds user to list of players
func (mydb *Sqldatabase) AddUser(u *models.User) error {
	db, err := sql.Open("mysql", mydb.SQLtoken)
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS Santa_Players(id INT NOT NULL AUTO_INCREMENT, userid INT NOT NULL, chatid BIGINT NOT NULL, firstname TEXT NOT NULL, lastname TEXT NOT NULL, middlename TEXT NOT NULL, city TEXT NOT NULL, nova_poshta INT NOT NULL, number BIGINT NOT NULL, gift_info TEXT NOT NULL, confirm BOOLEAN NOT NULL,PRIMARY KEY(id))")
	if err != nil {
		return err
	}

	_, err = db.Exec(
		"INSERT INTO Santa_Players(userid, chatid, firstname, lastname, middlename, city, nova_poshta, number, gift_info, confirm) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		u.UserID,
		u.ChatID,
		u.FirstName,
		u.LastName,
		u.MiddleName,
		u.City,
		u.NovaPoshta,
		u.Number,
		u.GiftInfo,
		u.Confirm,
	)
	if err != nil {
		return err
	}

	return nil

}

//AddUser this function adds user to list of players
func (mydb *Sqldatabase) UpdateUser(u *models.User) error {
	db, err := sql.Open("mysql", mydb.SQLtoken)
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS Santa_Players(id INT NOT NULL AUTO_INCREMENT, userid INT NOT NULL, chatid BIGINT NOT NULL, firstname TEXT NOT NULL, lastname TEXT NOT NULL, middlename TEXT NOT NULL, city TEXT NOT NULL, nova_poshta INT NOT NULL, number BIGINT NOT NULL, gift_info TEXT NOT NULL, confirm BOOLEAN NOT NULL,PRIMARY KEY(id))")
	if err != nil {
		return err
	}

	_, err = db.Exec(
		"UPDATE Santa_Players SET firstname = ?, lastname = ?, middlename = ?, city = ?, nova_poshta = ?, number = ?, gift_info = ?, confirm = ? WHERE userid = ? AND chatid = ?",
		u.FirstName,
		u.LastName,
		u.MiddleName,
		u.City,
		u.NovaPoshta,
		u.Number,
		u.GiftInfo,
		u.Confirm,
		u.UserID,
		u.ChatID,
	)
	if err != nil {
		return err
	}

	return nil

}

//RemoveUser removes user from list
func (mydb *Sqldatabase) RemoveUser(userID int, chatID int64) error {
	db, err := sql.Open("mysql", mydb.SQLtoken)
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS Santa_Players(id INT NOT NULL AUTO_INCREMENT, userid INT NOT NULL, chatid BIGINT NOT NULL, firstname TEXT NOT NULL, lastname TEXT NOT NULL, middlename TEXT NOT NULL, city TEXT NOT NULL, nova_poshta INT NOT NULL, number BIGINT NOT NULL, gift_info TEXT NOT NULL, confirm BOOLEAN NOT NULL,PRIMARY KEY(id))")
	if err != nil {
		return err
	}

	_, err = db.Exec("DELETE FROM Santa_Players WHERE userid = ? AND chatid = ?", userID, chatID)
	if err != nil {
		return err
	}
	return nil
}

//UserPresent checks if user is already in list
func (mydb *Sqldatabase) UserPresent(userID int, chatID int64) (*models.User, error) {
	db, err := sql.Open("mysql", mydb.SQLtoken)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS Santa_Players(id INT NOT NULL AUTO_INCREMENT, userid INT NOT NULL, chatid BIGINT NOT NULL, firstname TEXT NOT NULL, lastname TEXT NOT NULL, middlename TEXT NOT NULL, city TEXT NOT NULL, nova_poshta INT NOT NULL, number BIGINT NOT NULL, gift_info TEXT NOT NULL, confirm BOOLEAN NOT NULL,PRIMARY KEY(id))")
	if err != nil {
		return nil, err
	}

	u := new(models.User)
	err = db.QueryRow(
		"SELECT * FROM Santa_Players WHERE userid = ? AND chatid = ?", userID, chatID).Scan(
		&u.ID,
		&u.UserID,
		&u.ChatID,
		&u.FirstName,
		&u.LastName,
		&u.MiddleName,
		&u.City,
		&u.NovaPoshta,
		&u.Number,
		&u.GiftInfo,
		&u.Confirm,
	)
	if err != nil {
		return nil, err
	}

	return u, nil

}

//Confirm allows you to take part
func (mydb *Sqldatabase) Confirm(userID int, chatID int64) error {
	db, err := sql.Open("mysql", mydb.SQLtoken)
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS Santa_Players(id INT NOT NULL AUTO_INCREMENT, userid INT NOT NULL, chatid BIGINT NOT NULL, firstname TEXT NOT NULL, lastname TEXT NOT NULL, middlename TEXT NOT NULL, city TEXT NOT NULL, nova_poshta INT NOT NULL, number BIGINT NOT NULL, gift_info TEXT NOT NULL, confirm BOOLEAN NOT NULL,PRIMARY KEY(id))")
	if err != nil {
		return err
	}

	_, err = db.Exec("UPDATE Santa_Players SET confirm = true WHERE userid = ? AND chatid = ?", userID, chatID)

	if err != nil {
		return err
	}

	return nil

}

//SelectAllUsers gets all users from db
func (mydb *Sqldatabase) SelectAllUsers(chatID int64) ([]*models.User, error) {
	db, err := sql.Open("mysql", mydb.SQLtoken)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS Santa_Players(id INT NOT NULL AUTO_INCREMENT, userid INT NOT NULL, chatid BIGINT NOT NULL, firstname TEXT NOT NULL, lastname TEXT NOT NULL, middlename TEXT NOT NULL, city TEXT NOT NULL, nova_poshta INT NOT NULL, number BIGINT NOT NULL, gift_info TEXT NOT NULL, confirm BOOLEAN NOT NULL,PRIMARY KEY(id))")
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(
		"SELECT * FROM Santa_Players WHERE chatid = ?", chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []*models.User
	for rows.Next() {
		u := new(models.User)
		err := rows.Scan(
			&u.ID,
			&u.UserID,
			&u.ChatID,
			&u.FirstName,
			&u.LastName,
			&u.MiddleName,
			&u.City,
			&u.NovaPoshta,
			&u.Number,
			&u.GiftInfo,
			&u.Confirm,
		)
		if err != nil {
			fmt.Println("176 error sql")
		}
		users = append(users, u)
	}

	return users, nil

}

//GetUser get user from DB
func (mydb *Sqldatabase) GetUser(userID int, chatID int64) (*models.User, error) {
	db, err := sql.Open("mysql", mydb.SQLtoken)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS Santa_Players(id INT NOT NULL AUTO_INCREMENT, userid INT NOT NULL, chatid BIGINT NOT NULL, firstname TEXT NOT NULL, lastname TEXT NOT NULL, middlename TEXT NOT NULL, city TEXT NOT NULL, nova_poshta INT NOT NULL, number BIGINT NOT NULL, gift_info TEXT NOT NULL, confirm BOOLEAN NOT NULL,PRIMARY KEY(id))")
	if err != nil {
		return nil, err
	}
	u := &models.User{}
	err = db.QueryRow(
		"SELECT * FROM Santa_Players WHERE chatid = ? AND userid = ?", chatID, userID).Scan(
		&u.ID,
		&u.UserID,
		&u.ChatID,
		&u.FirstName,
		&u.LastName,
		&u.MiddleName,
		&u.City,
		&u.NovaPoshta,
		&u.Number,
		&u.GiftInfo,
		&u.Confirm,
	)
	if err != nil {
		return nil, err
	}

	return u, nil

}

//SelectAllUsersConfirmed gets all users from db
func (mydb *Sqldatabase) SelectAllUsersConfirmed(chatID int64) ([]*models.User, error) {
	db, err := sql.Open("mysql", mydb.SQLtoken)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS Santa_Players(id INT NOT NULL AUTO_INCREMENT, userid INT NOT NULL, chatid BIGINT NOT NULL, firstname TEXT NOT NULL, lastname TEXT NOT NULL, middlename TEXT NOT NULL, city TEXT NOT NULL, nova_poshta INT NOT NULL, number BIGINT NOT NULL, gift_info TEXT NOT NULL, confirm BOOLEAN NOT NULL,PRIMARY KEY(id))")
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(
		"SELECT * FROM Santa_Players WHERE chatid = ? AND confirm = true", chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []*models.User
	for rows.Next() {
		u := new(models.User)
		err := rows.Scan(
			&u.ID,
			&u.UserID,
			&u.ChatID,
			&u.FirstName,
			&u.LastName,
			&u.MiddleName,
			&u.City,
			&u.NovaPoshta,
			&u.Number,
			&u.GiftInfo,
			&u.Confirm,
		)
		if err != nil {
			fmt.Println("176 error sql")
		}
		users = append(users, u)
	}

	return users, nil

}

//SelectOneUserFromChooseSanta gets all users from db
func (mydb *Sqldatabase) SelectOneUserFromChooseSanta(userID int) (*models.ChooseSanta, error) {
	db, err := sql.Open("mysql", mydb.SQLtoken)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS Choose_Santa(id INT NOT NULL AUTO_INCREMENT, santa_for INT NOT NULL, santa_id INT NOT NULL, userid INT NOT NULL, chatid BIGINT NOT NULL, gift BOOLEAN NOT NULL, valid BOOLEAN NOT NULL, PRIMARY KEY(id))")
	if err != nil {
		return nil, err
	}

	u := new(models.ChooseSanta)
	err = db.QueryRow(
		"SELECT * FROM Choose_Santa WHERE userid = ?", userID).Scan(
			&u.ID,
			&u.SantaFor,
			&u.MySanta,
			&u.UserID,
			&u.ChatID,
			&u.Gift,
			&u.Valid,
		)
	if err != nil {
		return nil, err
	}

	return u, nil

}
//StartSelecting gets all users from db
func (mydb *Sqldatabase) StartSelecting(users []*models.User) ([]*models.ChooseSanta, error) {
	db, err := sql.Open("mysql", mydb.SQLtoken)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS Choose_Santa(id INT NOT NULL AUTO_INCREMENT, santa_for INT NOT NULL, santa_id INT NOT NULL, userid INT NOT NULL, chatid BIGINT NOT NULL, gift BOOLEAN NOT NULL, valid BOOLEAN NOT NULL, PRIMARY KEY(id))")
	if err != nil {
		return nil, err
	}

	var (
		newUsers     []*models.ChooseSanta //will be reused as final list
		tmpUsers     []*models.ChooseSanta
		canBeSanta   []*models.ChooseSanta = make([]*models.ChooseSanta, len(users))
		canHaveSanta []*models.ChooseSanta = make([]*models.ChooseSanta, len(users))
	)
	for i := range users {
		newUsers = append(newUsers, users[i].ToChoose())
	}

	copy(canBeSanta, newUsers)
	copy(canHaveSanta, newUsers)

	rand.Seed(time.Now().Unix())

	if len(canBeSanta) == 1 {
		return nil, errors.New("Not enough participants")
	} else if len(canBeSanta) == 2 {
		user1 := canBeSanta[0]
		user2 := canBeSanta[1]
		user2.MySanta = user1.UserID
		user2.SantaFor = user1.UserID
		user1.SantaFor = user2.UserID
		user1.MySanta = user2.UserID
		//change info about first user in canHaveSanta list
		tmpUsers = append(tmpUsers, user1, user2)
	} else {
	M:
		for len(canHaveSanta) != 0 {
			random1 := rand.Int() % len(canBeSanta)
			random2 := rand.Int() % len(canHaveSanta)
			user1 := canBeSanta[random1]
			user2 := canHaveSanta[random2]
			for user1.UserID == user2.UserID {
				random2 = rand.Int() % len(canHaveSanta)
				user2 = canHaveSanta[random2]

				if len(canBeSanta) == 1 && len(canHaveSanta) == 1 {
					canBeSanta = make([]*models.ChooseSanta, len(users))
					canHaveSanta = make([]*models.ChooseSanta, len(users))
					copy(canBeSanta, newUsers)
					copy(canHaveSanta, newUsers)
					tmpUsers = nil
					continue M
				}
			}
			user2.MySanta = user1.UserID
			user1.SantaFor = user2.UserID
			//if user1 is santa for user2 and user2 is santa for userid rerandom
			// for (user1.UserID == user2.MySanta && user2.UserID == user1.MySanta) || user1.UserID == user2.UserID {
			// 	random2 = rand.Int() % len(canHaveSanta)
			// 	user2 = canHaveSanta[random2]
			// }
			// user2.MySanta = user1.UserID
			// user1.SantaFor = user2.UserID

			//change info about first user in canHaveSanta list
			for i := range canHaveSanta {
				if user1.UserID == canHaveSanta[i].UserID {
					canHaveSanta[i].SantaFor = user2.UserID
					break
				}
			}
			tmpUsers = append(tmpUsers, user2)
			canHaveSanta = append(canHaveSanta[:random2], canHaveSanta[random2+1:]...)
			canBeSanta = append(canBeSanta[:random1], canBeSanta[random1+1:]...)
		}
	}

	for i := range tmpUsers {
		_, err := db.Exec("INSERT INTO Choose_Santa(santa_for, santa_id, userid, chatid, gift, valid) VALUES(?, ?, ?, ?, ?, ?)",
			tmpUsers[i].SantaFor,
			tmpUsers[i].MySanta,
			tmpUsers[i].UserID,
			tmpUsers[i].ChatID,
			tmpUsers[i].Gift,
			tmpUsers[i].Valid,
		)
		if err != nil {
			fmt.Println("SQL 321", err.Error())
		}
	}
	fmt.Println("finished roll")
	return tmpUsers, nil
}

// //DeleteOldRecords ...
// func (mydb *Sqldatabase) DeleteOldRecords() {
// 	db, err := sql.Open("mysql", mydb.SQLtoken)
// 	if err != nil {
// 		fmt.Println(err.Error())
// 		return
// 	}

// 	_, err := db.Exec("DELETE FROM Santa_Players WHERE ")
// }

//ChangePhotoStatus ...
func (mydb *Sqldatabase) ChangePhotoStatus(userID int, chatID int64) error {
	db, err := sql.Open("mysql", mydb.SQLtoken)
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS Choose_Santa(id INT NOT NULL AUTO_INCREMENT, santa_for INT NOT NULL, santa_id INT NOT NULL, userid INT NOT NULL, chatid BIGINT NOT NULL, gift BOOLEAN NOT NULL, valid BOOLEAN NOT NULL, PRIMARY KEY(id))")
	if err != nil {
		return err
	}
	_, err = db.Exec("UPDATE Choose_Santa SET valid = true, gift = true WHERE userid = ? AND chatid = ?", userID, chatID)
	if err != nil {
		return err
	}
	return nil
}

//GetUserFromChoosen ...
func (mydb *Sqldatabase) GetUserFromChoosen(userID int, chatID int64) (*models.ChooseSanta, error) {
	db, err := sql.Open("mysql", mydb.SQLtoken)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS Choose_Santa(id INT NOT NULL AUTO_INCREMENT, santa_for INT NOT NULL, santa_id INT NOT NULL, userid INT NOT NULL, chatid BIGINT NOT NULL, gift BOOLEAN NOT NULL, valid BOOLEAN NOT NULL, PRIMARY KEY(id))")
	if err != nil {
		return nil, err
	}

	u := new(models.ChooseSanta)
	err = db.QueryRow("SELECT * FROM Choose_Santa WHERE chatid = ? AND userid = ?", chatID, userID).Scan(
		&u.ID,
		&u.SantaFor,
		&u.MySanta,
		&u.UserID,
		&u.ChatID,
		&u.Gift,
		&u.Valid,
	)
	if err != nil {
		return nil, err
	}

	return u, nil
}

//ChangeGifted ...
func (mydb *Sqldatabase) ChangeGifted(u *models.ChooseSanta) (*models.ChooseSanta, error) {
	db, err := sql.Open("mysql", mydb.SQLtoken)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS Choose_Santa(id INT NOT NULL AUTO_INCREMENT, santa_for INT NOT NULL, santa_id INT NOT NULL, userid INT NOT NULL, chatid BIGINT NOT NULL, gift BOOLEAN NOT NULL, valid BOOLEAN NOT NULL, PRIMARY KEY(id))")
	if err != nil {
		return nil, err
	}

	err = db.QueryRow("UPDATE Choose_Santa SET santa_for = ? WHERE chatid = ? AND userid = ?", u.SantaFor, u.ChatID, u.MySanta).Scan(
		&u.ID,
		&u.SantaFor,
		&u.MySanta,
		&u.UserID,
		&u.ChatID,
		&u.Gift,
		&u.Valid,
	)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec("UPDATE Choose_Santa SET santa_id = ? WHERE userid = ? AND chatid = ?", u.MySanta, u.SantaFor, u.ChatID)
	if err != nil {
		return nil, err
	}
	return u, nil
}
