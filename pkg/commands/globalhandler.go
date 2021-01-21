package commands

import (
	"database/sql"
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

//GlobalMessages chan for global messages
var GlobalMessages chan *tgbotapi.Message

//GlobalHandler handler for global messages
func (command *CommandAPI) GlobalHandler() {
	for message := range GlobalMessages {
		switch message.Command() {
		case "start":
			go command.CheckPM(message)
			confirmUser <- message
		case "rules":
			command.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(message.Chat.ID, message.MessageID))
			confirmUser <- message
		case "play":
			go command.CheckPM(message)
			confirmUser <- message

		case "stop":
			go command.CheckPM(message)
			confirmUser <- message

		case "edit":
			go command.CheckPM(message)
			confirmUser <- message

		case "present":
			go command.CheckPM(message)
			confirmUser <- message	

		case "roll":
			command.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(message.Chat.ID, message.MessageID))
			if command.CheckAdmins(message) {
				command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Запущені всі алгоритми. Підтвердіть вашу участь у грі до півночі, або вас буде кікнуто. Потім з'явиться інформація про вашого щасливчика. Щоб підтвердити свою участь напишіть /confirm"))
				go command.InitRoll(message.Chat)
			} else {
				command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Не варто квапитись, адміни все зроблять, коли буде потрібно"))
			}
		case "confirm":
			go command.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(message.Chat.ID, message.MessageID))
			command.confirm(message)
		}
	}
}

//CheckPM message that requests to check private messages
func (command *CommandAPI) CheckPM(message *tgbotapi.Message) {
	command.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(message.Chat.ID, message.MessageID))
	msg, _ := command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Зазирни в пп"))
	time.Sleep(5 * time.Second)
	command.Bot.DeleteMessage(tgbotapi.NewDeleteMessage(msg.Chat.ID, msg.MessageID))
}

//CheckAdmins checks if current user is admin
func (command *CommandAPI) CheckAdmins(message *tgbotapi.Message) bool {
	chatConfigWithUser := tgbotapi.ChatConfigWithUser{
		ChatID: message.Chat.ID,
		UserID: message.From.ID,
	}
	member, err := command.Bot.GetChatMember(chatConfigWithUser)
	if err != nil {
		return false
	}
	if member.Status == "creator" || member.Status == "administrator" {
		return true
	}
	return false
}

func (command *CommandAPI) confirm(message *tgbotapi.Message) {
	u, err := command.DB.UserPresent(message.From.ID, message.Chat.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			command.Bot.Send(tgbotapi.NewMessage(int64(message.From.ID), "Ви не берете участі у грі. Якщо хочете додатись, пропишіть у чаті /play"))
			return
		}
	}
	
	if err := command.DB.Confirm(u.UserID, u.ChatID); err == nil {
		command.Bot.Send(tgbotapi.NewMessage(int64(message.From.ID), fmt.Sprintf("Ви підтвердили свою участь у %s", message.Chat.Title)))
		fmt.Printf("User with ID: %d has confirmed his participation\n", message.From.ID)
		return
	}
}
