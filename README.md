# PsDiscountsBot
Source code for telegram bot, that can parse discounted games from PlayStation Store and post this games to telegram channel.
![logo](https://github.com/GreyLabsDev/PsDiscountsBot/blob/master/playstation_logo.png)

## Main features
- Parsing game discounts page of PlayStation Store (package datasource)
- Writing discounted games into .json file by special data model (package models)
- Updating discounted games in .json file every day in defined time of day (packages taskmanager, telegram)
- Auto-posting of discounted games to any telegram channel (packages taskmanager, telegram)
- Manual posting control by special commends to bot (package telegram)
- Fully autonomus, all features scheduled with tasmanager package and executes at defined time every day

### Used technologies
* [go-telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api) - Golang bindings for the Telegram Bot API
* [phantomgo](https://github.com/k4s/phantomgo) - Rendering web-pages with .js components
* [GJSON](https://github.com/tidwall/gjson) - Fast and simple way to get values from .json files
