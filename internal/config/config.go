package config

import (
	"dursanta/pkg/commands"
	"dursanta/pkg/sqldatabase"
	"flag"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/BurntSushi/toml"
)

var (
	tokenToml  string
	parsedToml struct {
		BotToken string `toml:"bot_token"`
		SQLToken string `toml:"sql_token"`
	}
)

func init() {
	flag.StringVar(&tokenToml, "config-path", "./internal/config/config.toml", "path to config file")
}

//MakeConfig returns you new commandAPI
func MakeConfig() (*commands.CommandAPI, error) {
	flag.Parse()

	_, err := toml.DecodeFile(tokenToml, &parsedToml) //decode file
	if err != nil {
		panic("Broken config file " + err.Error())
	}
	//create bot instance
	bot, err := tgbotapi.NewBotAPI(parsedToml.BotToken)
	if err != nil {
		panic("Couldnt create bot " + err.Error())
	}
	//set commandAPI and Sqldatabase basic values
	commandAPI := new(commands.CommandAPI)
	commandAPI.DB = new(sqldatabase.Sqldatabase)
	commandAPI.Bot = bot
	commandAPI.DB.Bot = bot
	commandAPI.DB.SQLtoken = parsedToml.SQLToken
	//check mysql db
	if err := commandAPI.DB.CheckDB(); err != nil {
		panic("Cant open db " + err.Error())
	}

	return commandAPI, nil
}
