package main

import (
	"dursanta/internal/config"
	"dursanta/pkg/commands"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {
	commandAPI, err := config.MakeConfig()
	if err != nil {
		panic(err)
	}

	u := tgbotapi.UpdateConfig{
		Offset:  0,
		Timeout: 100,
	}

	updates, err := commandAPI.Bot.GetUpdatesChan(u)
	if err != nil {
		panic(err)
	}
	go initFunctions(commandAPI)
	//init global messages
	go commandAPI.GlobalHandler()
	//init private messages
	go commandAPI.PrivateHandler()

	processMessages(commandAPI, updates)
}
func initFunctions(commandAPI *commands.CommandAPI) {
	go commandAPI.DeleteOldRecords()

	//userIDs := []int64{933724087, 560120865, 765571765, 509938866, 722241703}
	// for _, v := range userIDs {
	// 	commandAPI.Bot.Send(tgbotapi.NewMessage(v, "Нагадую, що ви ще не підтвердили подарунок у чаті 'Мафія в бункері'. Нагадую умови\n\nТобі потрібно буде підтвердити наявність подарунку. Сфотографуй його (не забудь про чек з пошти), напиши /present у чаті та надійшли фото у боті."))
	// }
	// chatConfigwithUser := tgbotapi.ChatConfig{
	// 	ChatID: -418732502,
	// 	//UserID: 619754412,
	// }
	// chat, err := commandAPI.Bot.GetChat(chatConfigwithUser)
	// fmt.Println(chat, err)
	//commandAPI.Bot.Send(tgbotapi.NewMessage(int64(-418732502), "Підтримку цього бота незабаром буде закінчено(він створювався для одноразової гри у чаті, вказаному в описі)"))
}
func processMessages(commandAPI *commands.CommandAPI, updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		if update.Message == nil {
			continue
		}
		// if update.Message.Chat.ID != -1001349237738 {
		// 	fmt.Println(logMessage(update.Message))
		// }
		//message private
		if int64(update.Message.From.ID) == update.Message.Chat.ID {
			commands.PrivateMessages <- update.Message
		} else { // message global
			commands.GlobalMessages <- update.Message
		}


	}
}

func logMessage(message *tgbotapi.Message) string {
	return fmt.Sprintf("User\nID: %d\nFirstName: %s\nUserName: %s\nChat\nChatID: %d\nChatTitle: %s\nChatUser: %s\nInviteLink: %s\nText: %s\n", message.From.ID, message.From.FirstName, message.From.UserName, message.Chat.ID, message.Chat.Title, message.Chat.UserName, message.Chat.InviteLink, message.Text)
}

/*
TODO
1. IF USER WANTS TO LEAVE AFTER ROLL SHOULD REMOVE SANTA FOR AND SANTA ID FOR OTHERs
user1 santafor-> user2 santafor -> user3
user2 leaves
user1 santafor -> user3
Santa can be Santa for his Santa


2. Validity of photos check somehow and after that send full info.
3. Do not show all info before checking photo validity.
4. Clear all databases after 30 days. (add to rules text about deleting in 30 days)
(probably add date to database and check)
5. 0 at number start is not saved
(050 will be saved as 50)
6. Think about double roll started(and roll started and bot off)
(somebody rolled twice)
7. Remind of gift verification
(once in two days after roll finished)
*/
