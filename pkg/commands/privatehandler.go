package commands

import (
	"database/sql"
	"dursanta/internal/text"
	"dursanta/pkg/models"
	"dursanta/pkg/tools"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf16"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

//PrivateMessages chan for private messages
var (
	PrivateMessages chan *tgbotapi.Message
	confirmUser     chan *tgbotapi.Message
	closeChan       chan int
	doneChan        chan int
)

//PrivateHandler handler for global messages
func (command *CommandAPI) PrivateHandler() {
	type Safe struct {
		mu       *sync.Mutex
		userChan map[int]chan *tgbotapi.Message
	}

	safe := &Safe{
		mu:       &sync.Mutex{},
		userChan: make(map[int]chan *tgbotapi.Message),
	}

	for {
		select {
		//from main channel
		case message := <-PrivateMessages:
			if message.IsCommand() {
				switch message.Command() {
				case "start":
					go command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, text.StartMessage))
					continue
				case "rules":
					go command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, text.StartMessage))
					continue
				case "done":
					//do nothing here
				default:
					go command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Використовуй цю команду у чаті"))
					continue
				}
			}
			/*
			for key := range safe.userChan {
				if key == message.From.ID {
					safe.userChan[key] <- message
				}
			}
			*/
			safe.mu.Lock()
			if safe.userChan[message.From.ID] != nil {
				safe.userChan[message.From.ID] <- message
			}
			safe.mu.Unlock()

			//from others
		case message := <-confirmUser:
			switch message.Command() {
			case "start":
				go command.Bot.Send(tgbotapi.NewMessage(int64(message.From.ID), text.StartMessage))
				continue
			case "rules":
				go command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, text.StartMessage))
				continue
			case "play": //if user wants to register
				//CHECK IF USER IS ALREADY PLAYING IF YES ASK HIM TO USE EDIT
				u, err := command.DB.UserPresent(message.From.ID, message.Chat.ID)
				if err != nil {
					if err != sql.ErrNoRows {
						fmt.Println(err)
						command.Bot.Send(tgbotapi.NewMessage(int64(message.From.ID), "Щось пішло не так, напишіть @phantom_writer"))
						continue
					}
				}
				if u != nil {
					command.Bot.Send(tgbotapi.NewMessage(int64(message.From.ID), "Ти вже береш участь у грі. Якщо хочеш змінити інформацію про себе, використовуй /edit. Якщо хочеш припинити приймати участь, використовуй /stop"))
					continue
				}
				safe.mu.Lock()
				if safe.userChan[message.From.ID] != nil { //check if registration is not in process
					command.Bot.Send(tgbotapi.NewMessage(int64(message.From.ID), "Ви вже розпочали якусь дію"))
					safe.mu.Unlock()
					continue
				} else {
					//if user havent ever pressed start here should be error
					text := fmt.Sprintf("Привіт! Ти реєструєшся у грі «Таємний Санта» у чаті «%s»🎅🏻.\nДля участі та подальшій співпраці нам потрібно зібрати про тебе деяку інформацію😌\nПочнемо з твого імені. Як тебе звуть?(кирилицею)", message.Chat.Title)
					_, err := command.Bot.Send(tgbotapi.MessageConfig {
						Text: text,
						BaseChat: tgbotapi.BaseChat {
							ChatID: int64(message.From.ID),
							ReplyMarkup: tgbotapi.ReplyKeyboardHide{
								HideKeyboard: true,
							},
						},
					})
					if err != nil {
						fmt.Println(err.Error())
						safe.mu.Unlock()
						continue
					}
					//create chan
					safe.userChan[message.From.ID] = make(chan *tgbotapi.Message)
					go command.registerUser(safe.userChan[message.From.ID], message.From.ID, message.Chat.ID)
				}
				safe.mu.Unlock()

			case "stop":
				if _, err := command.DB.UserPresent(message.From.ID, message.Chat.ID); err != nil {
					if err == sql.ErrNoRows {
						command.Bot.Send(tgbotapi.NewMessage(int64(message.From.ID), "Ви не берете участі у грі. Якщо хочете додатись, пропишіть у чаті /play"))
						continue
					}
				}
				u, _ := command.DB.GetUserFromChoosen(message.From.ID, message.Chat.ID)
				if u != nil {
					command.Bot.Send(tgbotapi.NewMessage(int64(message.From.ID), "Ви не можете покинути поточну гру"))
					continue
				}
				//TODO
				//maybe here stop for people that already taking part
				safe.mu.Lock()
				if safe.userChan[message.From.ID] != nil { //check if registration is not in process
					command.Bot.Send(tgbotapi.NewMessage(int64(message.From.ID), "Ви вже розпочали якусь дію"))
					safe.mu.Unlock()
					continue
				} else {
					text := fmt.Sprintf("Ти точно хочеш вийти з гри у чаті %s?", message.Chat.Title)
					msg := tgbotapi.NewMessage(int64(message.From.ID), text)
					msg.BaseChat.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
						Keyboard: [][]tgbotapi.KeyboardButton{
							[]tgbotapi.KeyboardButton{
								tgbotapi.KeyboardButton{
									Text: "Так",
								},
								tgbotapi.KeyboardButton{
									Text: "Ні",
								},
							},
						},
						OneTimeKeyboard: true,
						ResizeKeyboard:  true,
					}
					_, err := command.Bot.Send(msg)
					if err != nil {
						fmt.Println(err.Error())
						safe.mu.Unlock()
						continue
					}
					//create chan
					safe.userChan[message.From.ID] = make(chan *tgbotapi.Message)
					go command.stopTakingPart(safe.userChan[message.From.ID], message.From.ID, message.Chat.ID)
				}
				safe.mu.Unlock()

			case "edit":
				u, err := command.DB.UserPresent(message.From.ID, message.Chat.ID)
				if err != nil {
					if err == sql.ErrNoRows {
						command.Bot.Send(tgbotapi.NewMessage(int64(message.From.ID), "Ви не берете участі у грі. Якщо хочете додатись, пропишіть у чаті /play"))
						continue
					}
				}

				safe.mu.Lock()
				if safe.userChan[message.From.ID] != nil { //check if registration is not in process
					command.Bot.Send(tgbotapi.NewMessage(int64(message.From.ID), "Ви вже розпочали якусь дію"))
					safe.mu.Unlock()
					continue
				} else {
					text := fmt.Sprintf("Ось ваша поточна інформація\n%s\nЩо саме ви хочете змінити", u)
					msg := tgbotapi.NewMessage(int64(message.From.ID), text)
					_, err := command.Bot.Send(msg)
					if err != nil {
						fmt.Println(err.Error())
						safe.mu.Unlock()
						continue
					}
					//create chan
					safe.userChan[message.From.ID] = make(chan *tgbotapi.Message)
					go command.editUser(safe.userChan[message.From.ID], message.From.ID, message.Chat.ID, u)
					command.changeInfo(message, int64(message.From.ID))
				}
				safe.mu.Unlock()

			case "present": //start sending photos
				u, _ := command.DB.GetUserFromChoosen(message.From.ID, message.Chat.ID)
				if u == nil {
					command.Bot.Send(tgbotapi.NewMessage(int64(message.From.ID), "Ви не берете участь у поточній грі."))
					continue
				}
				safe.mu.Lock()
				if safe.userChan[message.From.ID] != nil { //check if registration is not in process
					command.Bot.Send(tgbotapi.NewMessage(int64(message.From.ID), "Ви вже розпочали якусь дію"))
					safe.mu.Unlock()
					continue
				} else {
					text := "Надішліть скрін відправки подарунку новою поштою або чеку подарунку. Як тільки кинете всі фото, пропишіть /done"
					msg := tgbotapi.NewMessage(int64(message.From.ID), text)
					_, err := command.Bot.Send(msg)
					if err != nil {
						fmt.Println(err.Error())
						safe.mu.Unlock()
						continue
					}
					//create chan
					safe.userChan[message.From.ID] = make(chan *tgbotapi.Message)
					go command.confirmPresent(safe.userChan[message.From.ID], message.From.ID, message.Chat.ID)
				}
				safe.mu.Unlock()

			}

			//if users time is gone
		case userID := <-closeChan:
			safe.mu.Lock()
			close(safe.userChan[userID])
			delete(safe.userChan, userID)
			safe.mu.Unlock()
			//if user has finished his work OK
		case userID := <-doneChan:
			safe.mu.Lock()
			close(safe.userChan[userID])
			delete(safe.userChan, userID)
			safe.mu.Unlock()
		}
	}

}

func (command *CommandAPI) registerUser(messages chan *tgbotapi.Message, userID int, chatID int64) {
	fmt.Printf("User with ID %d started registration\n", userID)
	index := 0
	u := new(models.User)
	defer fmt.Printf("%s\nUser finished registration\n", u)
	u.UserID = userID
	u.ChatID = chatID
	u.Confirm = false
	timer := time.NewTicker(time.Minute * 7)
	editing := false
	for {
		select {
		case message := <-messages:
		i:
			timer = time.NewTicker(time.Minute * 7) //reset ticker
			switch index {
			case 0: //should enter name
				if err := tools.CheckName(utf16.Encode([]rune(message.Text))); err != nil {
					command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Неправильне ім'я"))
					continue
				}
				if message.Text != "" {
					u.FirstName = message.Text
				} else {
					command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Це поле не можна лишити порожнім"))
					continue
				}
				if editing {
					index = 8
					goto i
				}
				command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Тепер введи своє прізвище"))
				index++

			case 1: //should enter surname
				if err := tools.CheckName(utf16.Encode([]rune(message.Text))); err != nil {
					command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Неправильне прізвище"))
					continue
				}
				if message.Text != "" {
					u.LastName = message.Text
				} else {
					command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Це поле не можна лишити порожнім"))
					continue
				}
				if editing {
					index = 8
					goto i
				}
				command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Тепер введи по-батькові"))
				index++
			case 2: //should enter middlename
				if err := tools.CheckName(utf16.Encode([]rune(message.Text))); err != nil {
					command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Неправильне по-батькові"))
					continue
				}
				if message.Text != "" {
					u.MiddleName = message.Text
				} else {
					command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Це поле не можна лишити порожнім"))
					continue
				}
				if editing {
					index = 8
					goto i
				}
				command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Твій населений пункт."))
				index++
			case 3: //should enter city
				if message.Text != "" {
					u.City = message.Text
				} else {
					command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Це поле не можна лишити порожнім"))
					continue
				}
				if editing {
					index = 8
					goto i
				}
				command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Ах, чудове місце! Тепер введи відділення нової пошти, куди тобі прийде подаруночок."))
				index++

			case 4: //should enter nova_poshta viddil
				num, err := strconv.Atoi(message.Text)
				if err != nil {
					command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Введіть лише число"))
					continue
				}
				u.NovaPoshta = num
				if editing {
					index = 8
					goto i
				}
				command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Залишилось ще трохи! Введи свій номер телефону. Пам'ятай, вся введена інформація зашифрована навіть від розробника."))
				index++

			case 5: //sould enter phone_number
				num, err := strconv.ParseInt(message.Text, 10, 64)
				if err != nil {
					command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Введіть лише число"))
					continue
				}
				if message.Text[0] == '0' {
					num += 380000000000
				}
				u.Number = num
				if editing {
					index = 8
					goto i
				}
				command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "А тепер до найголовнішого🎅🏻\nРозкажи про свої вподобання, хобі, інтереси, щоб іншій людині було легше обрати для тебе подарунок, можеш кинути посилання на щось конкретне. Але пам'ятай, лише текстом + нагадуємо, що сума не має перевищувати 300-400 грн."))
				index++
			case 6: //any gift info
				if message.Text != "" {
					u.GiftInfo = message.Text
				} else {
					command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Це поле не можна лишити порожнім"))
					continue
				}
				if editing {
					index = 8
					goto i
				}
				msg := tgbotapi.MessageConfig{
					BaseChat: tgbotapi.BaseChat{
						ChatID: message.Chat.ID,
						ReplyMarkup: tgbotapi.ReplyKeyboardMarkup{
							Keyboard: [][]tgbotapi.KeyboardButton{
								[]tgbotapi.KeyboardButton{
									tgbotapi.KeyboardButton{
										Text: "Готово",
									},
									tgbotapi.KeyboardButton{
										Text: "Змінити",
									},
								},
							},
							OneTimeKeyboard: true,
							ResizeKeyboard:  true,
						},
					},
					Text: fmt.Sprintf("Перевірте введену вами інформацію.\n%s\nЯкщо ви з усім погоджуєтесь, натисніть готово. Якщо ні, то натисніть змінити.", u),
				}
				_, err := command.Bot.Send(msg)
				if err != nil {
					command.splitMessage(msg, msg.Text, err)
				}
				index++

			case 7: //should confirm info or try to edit something
				splittedMessage := strings.Split(message.Text, " ")
				switch splittedMessage[0] {
				case "Готово":
					doneChan <- userID
					err := command.DB.AddUser(u)
					if err != nil {
						fmt.Println(err)
					}
					timer.Stop()
					command.Bot.Send(tgbotapi.NewMessage(int64(userID), "Ви були зареєстровані!"))
					return
				case "Ні":
					message.Text = "Підтвердити"
					goto i
				case "Змінити":
					editing = true
					command.changeInfo(message, int64(userID))
				case "Так":
					command.changeInfo(message, int64(userID))
				case "Ім'я":
					index = 0
					command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Введіть ім'я"))
				case "Прізвище":
					index = 1
					command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Введіть прізвище"))
				case "По-батькові":
					index = 2
					command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Введіть по-батькові"))
				case "Місто":
					index = 3
					command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Введіть місто"))
				case "Нова":
					index = 4
					command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Введіть нову пошту"))
				case "Номер":
					index = 5
					command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Введіть номер"))
				case "Інформація":
					index = 6
					command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Введіть інформацію про подарунок"))
				case "Підтвердити":
					index = 7
					msg := tgbotapi.MessageConfig{
						BaseChat: tgbotapi.BaseChat{
							ChatID: message.Chat.ID,
							ReplyMarkup: tgbotapi.ReplyKeyboardMarkup{
								Keyboard: [][]tgbotapi.KeyboardButton{
									[]tgbotapi.KeyboardButton{
										tgbotapi.KeyboardButton{
											Text: "Готово",
										},
										tgbotapi.KeyboardButton{
											Text: "Змінити",
										},
									},
								},
								OneTimeKeyboard: true,
								ResizeKeyboard:  true,
							},
						},
						Text: fmt.Sprintf("Перевірте введену вами інформацію.\n%s\nЯкщо ви з усім погоджуєтесь, натисніть готово. Якщо ні, то натисніть змінити.", u),
					}
					_, err := command.Bot.Send(msg)
					if err != nil {
						command.splitMessage(msg, msg.Text, err)
					}
				}
				//send message about what do we want to edit
				//check for messagetext and do something. (change index due operation and )
			case 8:
				msg := tgbotapi.MessageConfig{
					BaseChat: tgbotapi.BaseChat{
						ChatID: message.Chat.ID,
						ReplyMarkup: tgbotapi.ReplyKeyboardMarkup{
							Keyboard: [][]tgbotapi.KeyboardButton{
								[]tgbotapi.KeyboardButton{
									tgbotapi.KeyboardButton{
										Text: "Так",
									},
									tgbotapi.KeyboardButton{
										Text: "Ні",
									},
								},
							},
							OneTimeKeyboard: true,
							ResizeKeyboard:  true,
						},
					},
					Text: fmt.Sprintf("Хочете змінити щось ще?"),
				}
				_, err := command.Bot.Send(msg)
				if err != nil {
					command.splitMessage(msg, msg.Text, err)
				}
				index = 7
			}

		case <-timer.C:
			timer.Stop()
			closeChan <- userID
			command.Bot.Send(tgbotapi.MessageConfig {
				Text: "Ваш час на реєстрацію був вичерпаний.",
				BaseChat: tgbotapi.BaseChat {
					ChatID: int64(userID),
					ReplyMarkup: tgbotapi.ReplyKeyboardHide{
						HideKeyboard: true,
					},
				},
			})
			fmt.Println("Finished unsuccessfully")
			return
		}
	}
}

func (command *CommandAPI) changeInfo(message *tgbotapi.Message, userID int64) {
	msg := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID: int64(userID),
			ReplyMarkup: tgbotapi.ReplyKeyboardMarkup{
				Keyboard: [][]tgbotapi.KeyboardButton{
					[]tgbotapi.KeyboardButton{
						tgbotapi.KeyboardButton{
							Text: "Ім'я",
						},
						tgbotapi.KeyboardButton{
							Text: "Прізвище",
						},
						tgbotapi.KeyboardButton{
							Text: "По-батькові",
						},
					},
					[]tgbotapi.KeyboardButton{
						tgbotapi.KeyboardButton{
							Text: "Місто",
						},
						tgbotapi.KeyboardButton{
							Text: "Нова пошта",
						},
						tgbotapi.KeyboardButton{
							Text: "Номер",
						},
					},
					[]tgbotapi.KeyboardButton{
						tgbotapi.KeyboardButton{
							Text: "Інформація про подарунок",
						},
						tgbotapi.KeyboardButton{
							Text: "Підтвердити",
						},
					},
				},
				OneTimeKeyboard: true,
				ResizeKeyboard:  true,
			},
		},
		Text: fmt.Sprintf("Оберіть те, що ви хочете змінити"),
	}
	_, err := command.Bot.Send(msg)
	if err != nil {
		command.splitMessage(msg, msg.Text, err)
	}
}

func (command *CommandAPI) stopTakingPart(messages chan *tgbotapi.Message, userID int, chatID int64) {
	fmt.Printf("User with ID %d wants to stop participating\n", userID)
	//TODO
	//Maybe add some logic when somebody leaves
	timer := time.NewTicker(time.Minute * 7)
	for {
		select {
		case message := <-messages:
			splittedMessage := strings.Split(message.Text, " ")
			switch splittedMessage[0] {
			case "Так":
				doneChan <- userID
				err := command.DB.RemoveUser(userID, chatID)
				if err != nil {
					fmt.Println(err)
				}
				// //this part is for moment when roll is already done
				// u, err := command.DB.GetUserFromChoosen(userID, chatID)
				// if err == nil {
				// 	//change gifted and santa
				// 	newGifted, err := command.DB.ChangeGifted(u)
				// 	if err != nil {
				// 		fmt.Println(err)
				// 	}

				// 	if newGifted != nil {
				// 		newUser, err := command.DB.GetUser(newGifted.SantaFor, newGifted.ChatID)
				// 		if err != nil {
				// 			fmt.Println(err)
				// 		}
				// 		/*SEND MESSAGE TO NEW SANTA*/
				// 		text := fmt.Sprintf("Твій попередній одержувач вийшов з гри. Ось твій новий учасник:\n")
				// 		msg := tgbotapi.MessageConfig{
				// 			BaseChat: tgbotapi.BaseChat{
				// 			},
				// 			ParseMode: "markdown",
				// 		}

				// 		msg.BaseChat.ChatID = int64(newGifted.UserID)

				// 		msg.Text = text + fmt.Sprintf("ПІБ:\n%s %s %s\n\nІнформація для відправлення:\nМісто: %s\nНомер відділення нової пошти: %d\nНомер телефону: %d\n\nВподобання/інтереси:\n%s\n\nТобі потрібно буде підтвердити наявність подарунку. Сфотографуй його (не забудь про чек з пошти), напиши /present у чаті та надійшли фото у боті.",
				// 			newUser.LastName,
				// 			newUser.FirstName,
				// 			newUser.MiddleName,
				// 			newUser.City,
				// 			newUser.NovaPoshta,
				// 			newUser.Number,
				// 			newUser.GiftInfo,)
				// 		_, err = command.Bot.Send(msg)
				// 		if err != nil {
				// 			command.splitMessage(msg, msg.Text, err)
				// 		}

				// 	}

				// }
				timer.Stop()
				command.Bot.Send(tgbotapi.NewMessage(int64(userID), "Ви більше не берете участі у грі."))
				return
			case "Ні":
				doneChan <- userID
				timer.Stop()
				command.Bot.Send(tgbotapi.NewMessage(int64(userID), "Все чудово!"))
				return
			}
		case <-timer.C:
			timer.Stop()
			closeChan <- userID
			command.Bot.Send(tgbotapi.MessageConfig {
				Text: "Ваш час прийняття рішення було вичерпано",
				BaseChat: tgbotapi.BaseChat {
					ChatID: int64(userID),
					ReplyMarkup: tgbotapi.ReplyKeyboardHide{
						HideKeyboard: true,
					},
				},
			})
			fmt.Println("Finished unsuccessfully")
			return
		}
	}
}

func (command *CommandAPI) editUser(messages chan *tgbotapi.Message, userID int, chatID int64, u *models.User) {
	fmt.Printf("User with ID %d started editing\n", userID)
	defer fmt.Printf("%s\nUser finished editing\n", u)

	index := 7
	timer := time.NewTicker(time.Minute * 7)

	for {
		select {
		case message := <-messages:
		i:
			timer = time.NewTicker(time.Minute * 7) //reset ticker
			switch index {
			case 0: //should enter name
				if err := tools.CheckName(utf16.Encode([]rune(message.Text))); err != nil {
					command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Неправильне ім'я"))
					continue
				}
				if message.Text != "" {
					u.FirstName = message.Text
				} else {
					command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Це поле не можна лишити порожнім"))
					continue
				}
				index = 8
				goto i

			case 1: //should enter surname
				if err := tools.CheckName(utf16.Encode([]rune(message.Text))); err != nil {
					command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Неправильне прізвище"))
					continue
				}
				if message.Text != "" {
					u.LastName = message.Text
				} else {
					command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Це поле не можна лишити порожнім"))
					continue
				}
				index = 8
				goto i

			case 2: //should enter middlename
				if err := tools.CheckName(utf16.Encode([]rune(message.Text))); err != nil {
					command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Неправильне по-батькові"))
					continue
				}
				if message.Text != "" {
					u.MiddleName = message.Text
				} else {
					command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Це поле не можна лишити порожнім"))
					continue
				}
				index = 8
				goto i

			case 3: //should enter city
				if message.Text != "" {
					u.City = message.Text
				} else {
					command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Це поле не можна лишити порожнім"))
					continue
				}
				index = 8
				goto i

			case 4: //should enter nova_poshta viddil
				num, err := strconv.Atoi(message.Text)
				if err != nil {
					command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Введіть лише число"))
					continue
				}
				u.NovaPoshta = num
				index = 8
				goto i

			case 5: //sould enter phone_number
				num, err := strconv.ParseInt(message.Text, 10, 64)
				if err != nil {
					command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Введіть лише число"))
					continue
				}
				if message.Text[0] == '0' {
					num += 380000000000
				}
				u.Number = num
				index = 8
				goto i

			case 6: //any gift info
				if message.Text != "" {
					u.GiftInfo = message.Text
				} else {
					command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Це поле не можна лишити порожнім"))
					continue
				}
				index = 8
				goto i

			case 7: //should confirm info or try to edit something
				splittedMessage := strings.Split(message.Text, " ")
				switch splittedMessage[0] {
				case "Готово":
					doneChan <- userID
					err := command.DB.UpdateUser(u)
					if err != nil {
						fmt.Println(err)
					}
					timer.Stop()
					command.Bot.Send(tgbotapi.NewMessage(int64(userID), "Ваша інформація була змінена"))
					return
				case "Ні":
					message.Text = "Підтвердити"
					goto i
				case "Змінити":
					command.changeInfo(message, int64(userID))
				case "Так":
					command.changeInfo(message, int64(userID))
				case "Ім'я":
					index = 0
					command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Введіть ім'я"))
				case "Прізвище":
					index = 1
					command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Введіть прізвище"))
				case "По-батькові":
					index = 2
					command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Введіть по-батькові"))
				case "Місто":
					index = 3
					command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Введіть місто"))
				case "Нова":
					index = 4
					command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Введіть нову пошту"))
				case "Номер":
					index = 5
					command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Введіть номер"))
				case "Інформація":
					index = 6
					command.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Введіть інформацію про подарунок"))
				case "Підтвердити":
					index = 7
					msg := tgbotapi.MessageConfig{
						BaseChat: tgbotapi.BaseChat{
							ChatID: message.Chat.ID,
							ReplyMarkup: tgbotapi.ReplyKeyboardMarkup{
								Keyboard: [][]tgbotapi.KeyboardButton{
									[]tgbotapi.KeyboardButton{
										tgbotapi.KeyboardButton{
											Text: "Готово",
										},
										tgbotapi.KeyboardButton{
											Text: "Змінити",
										},
									},
								},
								OneTimeKeyboard: true,
								ResizeKeyboard:  true,
							},
						},
						Text: fmt.Sprintf("Перевірте введену вами інформацію.\n%s\nЯкщо ви з усім погоджуєтесь, натисніть готово. Якщо ні, то натисніть змінити.", u),
					}
					_, err := command.Bot.Send(msg)
					if err != nil {
						command.splitMessage(msg, msg.Text, err)
					}
				}
				//send message about what do we want to edit
				//check for messagetext and do something. (change index due operation and )
			case 8:
				msg := tgbotapi.MessageConfig{
					BaseChat: tgbotapi.BaseChat{
						ChatID: message.Chat.ID,
						ReplyMarkup: tgbotapi.ReplyKeyboardMarkup{
							Keyboard: [][]tgbotapi.KeyboardButton{
								[]tgbotapi.KeyboardButton{
									tgbotapi.KeyboardButton{
										Text: "Так",
									},
									tgbotapi.KeyboardButton{
										Text: "Ні",
									},
								},
							},
							OneTimeKeyboard: true,
							ResizeKeyboard:  true,
						},
					},
					Text: fmt.Sprintf("Хочете змінити щось ще?"),
				}
				_, err := command.Bot.Send(msg)
				if err != nil {
					command.splitMessage(msg, msg.Text, err)
				}
				index = 7
			}

		case <-timer.C:
			timer.Stop()
			closeChan <- userID
			command.Bot.Send(tgbotapi.MessageConfig {
				Text: "Ваш час на зміну даних був вичерпаний.",
				BaseChat: tgbotapi.BaseChat {
					ChatID: int64(userID),
					ReplyMarkup: tgbotapi.ReplyKeyboardHide{
						HideKeyboard: true,
					},
				},
			})
			fmt.Println("Finished unsuccessfully")
			return
		}
	}
}

func (command *CommandAPI) confirmPresent(messages chan *tgbotapi.Message, userID int, chatID int64) {
	var messageIDS []int
	timer := time.NewTicker(7 * time.Minute)
	for {
		select {
		case message := <-messages:
			timer = time.NewTicker(7 * time.Minute)
			switch message.Command() {
			case "done":
				doneChan <- userID
				command.Bot.Send(tgbotapi.NewMessage(-1001177745465, fmt.Sprintf("Фото від %d", userID)))
				for _, messageID := range messageIDS {
					command.Bot.Send(tgbotapi.NewForward(-1001177745465, int64(userID), messageID))
				}
				return
			}
			messageIDS = append(messageIDS, message.MessageID)

		case <-timer.C:
			timer.Stop()
			closeChan <- userID
			command.Bot.Send(tgbotapi.MessageConfig {
				Text: "Ваш час на надсилання фото був вичерпаний",
				BaseChat: tgbotapi.BaseChat {
					ChatID: int64(userID),
					ReplyMarkup: tgbotapi.ReplyKeyboardHide{
						HideKeyboard: true,
					},
				},
			})
			fmt.Println("Finished unsuccessfully")
			return
		}
	}

}
