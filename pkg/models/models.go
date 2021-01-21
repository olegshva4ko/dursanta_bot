package models

import "fmt"

//User user in db
type User struct {
	ID         int    `sql:"id"`
	UserID     int    `sql:"userid"`
	ChatID     int64  `sql:"chatid"`
	FirstName  string `sql:"firstname"`
	LastName   string `sql:"lastname"`
	MiddleName string `sql:"middlename"`
	City       string `sql:"city"`
	NovaPoshta int    `sql:"nova_poshta"`
	Number     int64  `sql:"number"`
	GiftInfo   string `sql:"gift_info"` //information about dreamed gift
	Confirm    bool   `sql:"confirm"`   //information if user takes part in game
}

func (u *User) String() string {
	return fmt.Sprintf("Ваше ПІБ:\n%s %s %s\nІнформація про місцезнаходження:\n%s\nНомер відділення нової пошти:\n%d\nВаш номер:\n%d\nІнформація про вас/подарунок:\n%s",
		u.FirstName,
		u.LastName,
		u.MiddleName,
		u.City,
		u.NovaPoshta,
		u.Number,
		u.GiftInfo)
}

//ToChoose ...
func (u *User) ToChoose() *ChooseSanta {
	return &ChooseSanta{
		SantaFor: 0,
		MySanta:  0,
		UserID:   u.UserID,
		ChatID:   u.ChatID,
		Gift:     false,
		Valid:    false,
	}
}

//ChooseSanta user in db
type ChooseSanta struct {
	ID       int   `sql:"id"`
	SantaFor int   `sql:"santa_for"`
	MySanta  int   `sql:"santa_id"`
	UserID   int   `sql:"userid"`
	ChatID   int64 `sql:"chatid"`
	Gift     bool  `sql:"gift"` //ready or not
	Valid    bool  `sql:"valid"`
}


func (c *ChooseSanta) String() string {
	s := fmt.Sprintf("\n---\nID IN DB: %d\n", c.ID)
	s += fmt.Sprintf("UserID is: %d\n", c.UserID)
	if c.SantaFor != 0 {
		s += fmt.Sprintf("He is santa for user with userID: %d\n", c.SantaFor)
	}
	if c.MySanta != 0 {
		s += fmt.Sprintf("He has santa with userID: %d\n---\n", c.MySanta)
	}
	return s
}
// //ChooseSantaIterator ...
// type ChooseSantaIterator struct {
// 	Users        []*ChooseSanta //main array of elements
// 	CanBeSanta   []*ChooseSanta //users that already have a santa
// 	CanHaveSanta []*ChooseSanta //users that already are a santas
// 	Curr         int            //current element
// 	InARow       int            //count how many users with santa happend
// }

// //Next returns next element and go on circle
// func (c *ChooseSantaIterator) Next() *ChooseSanta {
// 	c.Curr++
// 	if c.Curr == len(c.Users) {
// 		c.Curr = 0
// 	}
// 	if c.Users[c.Curr].MySanta != 0 {
// 		c.InARow++
// 	}
// 	return c.Users[c.Curr]
// }

// // //NextRandom ...
// // func (c *ChooseSantaIterator) NextRandom() *ChooseSanta {

// // }
