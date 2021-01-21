package commands

import (
	"dursanta/pkg/sqldatabase"
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

//CommandAPI allows to use commands
type CommandAPI struct {
	Bot *tgbotapi.BotAPI
	DB  *sqldatabase.Sqldatabase
}

func init() {
	GlobalMessages = make(chan *tgbotapi.Message)
	PrivateMessages = make(chan *tgbotapi.Message)
	confirmUser = make(chan *tgbotapi.Message)
	closeChan = make(chan int)
	doneChan = make(chan int)
}

//splitMessage cuts text into smaller peaces if error appeared
func (command *CommandAPI) splitMessage(msg tgbotapi.MessageConfig, text string, err error) {
	if err.Error() == "Bad Request: message is too long" {
		howManyPieces := len(text) / 4096
		for i := 0; i <= howManyPieces; i++ {
			if i == howManyPieces {
				msg.Text = text[i*4096:]
				command.Bot.Send(msg)
				return
			}
			msg.Text = text[i*4096 : (i+1)*4096]
			command.Bot.Send(msg)

		}
	} else if string(err.Error()[0:34]) == "Bad Request: can't parse entities:" {
		splittedError := strings.Split(err.Error(), " ")
		index, err := strconv.Atoi(splittedError[len(splittedError)-1]) //last symbol is offset of unclosed entity
		if err != nil {
			return
		}
		howManyPieces := len(text) / (index - 1)

		for i := 0; i <= howManyPieces; i++ {
			if i == howManyPieces {
				msg.Text = text[i*(index-1):]
				command.Bot.Send(msg)
				return
			}
			msg.Text = text[i*(index-1) : (i+1)*(index-1)]
			command.Bot.Send(msg)
		}
	} else if len(err.Error()) > 20 && string(err.Error()[0:18]) == "Too Many Requests:" {
		timeToSleep := strings.Split(err.Error(), " ")
		parseTime, _ := strconv.Atoi(timeToSleep[len(timeToSleep)-1])
		time.Sleep(time.Duration(parseTime) * time.Millisecond)
		command.Bot.Send(msg)
	}
}


//InitRoll sleeps til midnight and starts send messages
func (command *CommandAPI) InitRoll(chat *tgbotapi.Chat) {
	fmt.Println("Starting roll...")
	//ADD MESSAGES SEND FOR CONFIRMATION
	msg := tgbotapi.MessageConfig{
		Text:      fmt.Sprintf("Підтверди свою участь у грі «Таємний Санта» у чаті «[%s](%s)»🎅🏻\nЗайди у чат та пропиши /confirm", chat.Title, chat.InviteLink),
		BaseChat:  tgbotapi.BaseChat{},
		ParseMode: "markdown",
	}
	users, err := command.DB.SelectAllUsers(chat.ID)
	if err != nil {
		fmt.Println("76 commands")
	}
	for _, user := range users {
		msg.BaseChat.ChatID = int64(user.UserID)
		_, err := command.Bot.Send(msg)
		if err != nil {
			command.splitMessage(msg, msg.Text, err)
		}
		time.Sleep(time.Millisecond * 300)
	}
	loc, _ := time.LoadLocation("Local")
	currTime := time.Now().Add(time.Hour * 24) //go to next date
	startTime := time.Date(                    //truncate it to 00 00
		currTime.Year(),
		currTime.Month(),
		currTime.Day(),
		0, 0, 0, 0, loc)

	time.Sleep(startTime.Sub(time.Now())) //start at midnight
	users, err = command.DB.SelectAllUsersConfirmed(chat.ID)
	if err != nil {
		fmt.Println("98 commands")
	}
	newUsers, err := command.DB.StartSelecting(users)
	if err != nil {
		fmt.Println(err.Error())
		if err.Error() == "Not enough participants" {
			command.Bot.Send(tgbotapi.NewMessage(chat.ID, "У цьому чаті недостатньо людей для початку."))
			return
		}
	}

	text := fmt.Sprintf("Привіт. Ти — таємний Санта для цієї людини:\n")
	msg = tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{},
	}

M:
	for i := range newUsers {
		for j := range users {
			if newUsers[i].SantaFor == users[j].UserID {
				msg.BaseChat.ChatID = int64(newUsers[i].UserID)

				msg.Text = text + fmt.Sprintf("ПІБ:\n%s %s %s\n\nІнформація для відправлення:\nМісто: %s\nНомер відділення нової пошти: %d\nНомер телефону: %d\n\nВподобання/інтереси:\n%s\n\nТобі потрібно буде підтвердити наявність подарунку. Сфотографуй його (не забудь про чек з пошти), напиши /present у чаті та надійшли фото у боті.",
					users[j].LastName,
					users[j].FirstName,
					users[j].MiddleName,
					users[j].City,
					users[j].NovaPoshta,
					users[j].Number,
					users[j].GiftInfo)
				_, err := command.Bot.Send(msg)
				if err != nil {
					fmt.Printf("Error while sending message to userID: %d\n", newUsers[i].UserID)
					command.splitMessage(msg, msg.Text, err)
				}
				time.Sleep(time.Millisecond * 300)
				continue M
			}
		}

	}
}

//DeleteOldRecords deletes old records from sql(more than 30 days)
func (command *CommandAPI) DeleteOldRecords() {
	// for {
	// 	command.DB.DeleteOldRecords()
	// 	time.Sleep(24 * time.Hour)
	// }
	/*
		TRY TO DELETE OLD RECORDS(OLDER THAN 30 DAYS) TO AVOID SPAMMING WITH NEW ROLLS
		AND POSSIBLE ERRORS OF SENDING TWO OR MORE USERS
	*/
}
