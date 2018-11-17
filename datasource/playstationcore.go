package datasource

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	filetool "psDicountsBot/file"
	"psDicountsBot/logger"
	. "psDicountsBot/models"
	stringtool "psDicountsBot/stringtool"
	"psDicountsBot/taskmanager"

	. "github.com/k4s/phantomgo"
	"github.com/tidwall/gjson"
)

var TagPS = "PlayStation."

var gameSourcePs = "store.playstation.com"
var gameIdPrefixPs = "PS_"

var baseUrlPs = "https://store.playstation.com"
var baseUrlSalesPs = "https://store.playstation.com/ru-ru/grid/STORE-MSF75508-PRICEDROPSCHI/"
var baseUrlSalesPsPostfix = "?gameContentType=games"

var gamesForPublicationPs []GameGeneral

var activeRepositoryPs = "gameListPS_0.json"
var proxyIsOnStatusPs = false

func PsCoreTest() {
	fmt.Println("PS core test")
}

func PsSetActiveRepository(path string) {
	activeRepositoryPs = path
}

func PsSwitchProxy(status bool) {
	proxyIsOnStatusPs = status
}

func PsGetInitForPublicationTask() taskmanager.SingleTask {
	return func() {
		PsInitForPublication()
	}
}

func PsGetUpdateDiscountedGamesTask() taskmanager.SingleTask {
	return func() {
		newGames := PsParseDiscountedGames()
		PsMegreAndSaveGamesToRepository(newGames)
	}
}

func PsInitForPublication() {
	gamesForPublicationPs, _ = PsLoadGamesFromRepo("notYetPublished", 0, 0)
	fmt.Println("PsInitForPublication.Inited - " + strconv.Itoa(len(gamesForPublicationPs)) + " games ready.")
	logger.Write("PsInitForPublication.Inited - " + strconv.Itoa(len(gamesForPublicationPs)) + " games ready.")
}

func PsGetRandomDiscountedGame() (game GameGeneral) {
	if len(gamesForPublicationPs) == 0 {
		return
	}

	i := randomPs(0, len(gamesForPublicationPs))
	game = gamesForPublicationPs[i]

	copy(gamesForPublicationPs[i:], gamesForPublicationPs[i+1:])
	gamesForPublicationPs = gamesForPublicationPs[:len(gamesForPublicationPs)-1]
	PsUpdateGameStatusInRepo(game.GlobalID, true)

	return game
}

func PsUpdateGameStatusInRepo(globalId string, isAlreadyPublished bool) {
	gamesFromRepo, isRepoExist := PsLoadGamesFromRepo("all", 0, 0)
	if !isRepoExist {
		return
	}
	gameIndex := GetIndexByGlobalId(gamesFromRepo, globalId)
	if gameIndex >= 0 {
		gamesFromRepo[gameIndex].AlreadyPublished = isAlreadyPublished
		PsSaveGamesToRepository(gamesFromRepo)
	} else {
		return
	}
}

func PsMegreAndSaveGamesToRepository(newGames []GameGeneral) {
	oldGames, isRepoFileExist := PsLoadGamesFromRepo("all", 0, 0)
	if !isRepoFileExist {
		PsSaveGamesToRepository(newGames)
		fmt.Println("no games in repo")
	} else {
		fmt.Println("found games in repo")
		mergedGames := MergeGameLists(oldGames, newGames)
		PsSaveGamesToRepository(mergedGames)
	}
}

func PsParseDiscountedGames() (games []GameGeneral) {
	startPageNumber := 1
	fmt.Println("PlayStationCore.PsParseDiscountedGames")
	p := &Param{
		Method:       "GET",
		Url:          baseUrlSalesPs + strconv.Itoa(startPageNumber) + baseUrlSalesPsPostfix,
		UsePhantomJS: true,
	}
	browser := NewPhantom()

	if proxyIsOnStatusPs {
		browser.SetProxyType("http")
		browser.SetProxyAuth("your_proxy_login:password")
		browser.SetProxy("your_proxy_ip:port")
	}

	resp, err := browser.Download(p)
	if err != nil {
		fmt.Println(err)
	}
	pageBody, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	pageBodyString := string(pageBody)
	pagesCount := psGetDiscountPagesCount(pageBodyString)

	pageBodyString = stringtool.ExtractBetween("grid-cell-container", "grid-footer-controls", pageBodyString)
	gamesRawDataArray := strings.Split(pageBodyString, "class=\"grid-cell grid-cell--game\">")
	gamesRawDataArray = append(gamesRawDataArray[:0], gamesRawDataArray[0+1:]...)
	fmt.Println("PlayStationCore.PsParseDiscountedGames - Found game pages = " + strconv.Itoa(pagesCount))
	loadedGames, idCounter := psExtractGamesFromRawArray(gamesRawDataArray, 0)
	fmt.Println("PlayStationCore.PsParseDiscountedGames - Found game pages = " + strconv.Itoa(len(loadedGames)))
	games = append(games, loadedGames...)

	if startPageNumber < pagesCount {
		counter := startPageNumber + 1
		for counter <= pagesCount {
			newParam := &Param{
				Method:       "GET",
				Url:          baseUrlSalesPs + strconv.Itoa(counter) + baseUrlSalesPsPostfix,
				UsePhantomJS: true,
			}
			resp, err := browser.Download(newParam)
			if err != nil {
				fmt.Println(err)
			}
			pageBody, err := ioutil.ReadAll(resp.Body)
			defer resp.Body.Close()

			pageBodyString := string(pageBody)
			pageBodyString = stringtool.ExtractBetween("grid-cell-container", "grid-footer-controls", pageBodyString)
			gamesRawDataArray := strings.Split(pageBodyString, "class=\"grid-cell grid-cell--game\">")
			gamesRawDataArray = append(gamesRawDataArray[:0], gamesRawDataArray[0+1:]...)
			// fmt.Println("GogCore.GogParseDiscountedGames - Found games = " + strconv.Itoa(len(gamesRawDataArray)))

			newGames, nextIdCounter := psExtractGamesFromRawArray(gamesRawDataArray, idCounter)
			idCounter = nextIdCounter
			games = append(games, newGames...)

			counter++
		}
	} else {
		return games
	}
	return games
}

func psExtractGamesFromRawArray(gamesRawDataArray []string, startCounter int) (games []GameGeneral, endCounter int) {
	counter := 0
	if startCounter != 0 {
		counter = startCounter
	}
	for _, rawGame := range gamesRawDataArray {
		re := regexp.MustCompile("[0-9]+")

		gameName := stringtool.ExtractBetween("class=\"grid-cell__title \">", "<div class=\"grid-cell__bottom\">", rawGame)
		gameName = strings.Split(gameName, "</span>")[0]
		gameName = strings.Split(gameName, ">")[1]

		gamePriceRaw := stringtool.ExtractBetween("\"price-display__price\">", "button data-fastboot-event-queue=\"add-to-cart", rawGame)
		gamePriceRaw = strings.Split(gamePriceRaw, "</h3>")[0]
		gamePriceRaw = strings.Replace(gamePriceRaw, ".", "", -1)
		if len(re.FindAllString(gamePriceRaw, 1)) >= 1 {
			gamePriceRaw = re.FindAllString(gamePriceRaw, 1)[0]
		} else {
			gamePriceRaw = "0"
		}

		gamePrice, _ := strconv.ParseInt(gamePriceRaw, 10, 14)

		gameDiscountString := stringtool.ExtractBetween("<span class=\"discount-badge__message\">", "<div class=\"grid-cell__body\">", rawGame)
		gameDiscountString = strings.Split(gameDiscountString, "</span>")[0]
		var err error
		var gameDiscount int64

		if strings.Contains(gameDiscountString, "РАСПРОД") {
			gameOldPriceRaw := stringtool.ExtractBetween("<span class=\"price-display__strikethrough\">", "class=\"price-display__price\"", rawGame)
			gameOldPriceRaw = stringtool.ExtractBetween("class=\"price\">", "</div>", gameOldPriceRaw)
			gameOldPriceRaw = strings.Replace(gameOldPriceRaw, ".", "", -1)
			gameOldPriceRaw = re.FindAllString(gameOldPriceRaw, 1)[0]
			oldPrice, parseErr := strconv.ParseInt(gameOldPriceRaw, 10, 14)
			if parseErr != nil {
				gameDiscount = 0
			} else {
				diff := float64(oldPrice - gamePrice)
				gameDiscount = int64((diff / float64(oldPrice)) * 100)
			}
		} else {
			if gameDiscountString == "" {
				gameDiscount = 0
			} else {
				gameDiscountString = re.FindAllString(gameDiscountString, 1)[0]
				gameDiscount, err = strconv.ParseInt(gameDiscountString, 10, 14)
				if err != nil {
					gameDiscount = 0
				}
			}

		}
		gameLinkRaw := strings.Split(rawGame, "grid-cell__prices-container\">")[1]
		gameLinkRaw = strings.Split(rawGame, "class=\"internal-app-link ember-view\">")[0]
		gameLinkRaw = stringtool.ExtractBetween("href=", "id", gameLinkRaw)
		gameLinkRaw = strings.Replace(gameLinkRaw, " ", "", -1)
		gameLinkRaw = strings.Replace(gameLinkRaw, "\"", "", -1)
		gameLink := gameLinkRaw
		gameHeaderRaw := stringtool.ExtractBetween("product-image__img product-image__img--main\">", "<div class=\"product-image__discount-badge\">", rawGame)
		gameHeaderRaw = stringtool.ExtractBetween("3x", "4x", gameHeaderRaw)
		gameHeaderRaw = strings.Replace(gameHeaderRaw, " ", "", -1)
		gameHeaderRaw = strings.Replace(gameHeaderRaw, ",", "", 1)
		// fmt.Println("Header 1st " + gameHeaderRaw)
		gameHeaderRaw = strings.Replace(gameHeaderRaw, "\u0026amp;", "&", -1)
		// fmt.Println("Header 2nd " + gameHeaderRaw)
		games = append(games, GameGeneral{
			gameIdPrefixPs + strconv.Itoa(counter),
			gameName,
			gameHeaderRaw,
			false,
			gameDiscount,
			gamePrice,
			gameLink,
			gameSourcePs,
			0,
			0,
			false,
			[]int{},
			false,
		})
		fmt.Println("Extracted game : " + gameName)
		counter++
	}
	endCounter = counter
	return games, endCounter
}

func psGetDiscountPagesCount(rawPageBody string) (count int) {
	if strings.Contains(rawPageBody, "paginator-control__end paginator-control__arrow-navigation") {
		var totalPagesString = strings.Split(rawPageBody, "paginator-control__end paginator-control__arrow-navigation")[0]
		tmpArray := strings.Split(totalPagesString, "grid/STORE-MSF75508-PRICEDROPSCHI")
		totalPagesString = tmpArray[len(tmpArray)-1]
		totalPagesString = strings.Split(totalPagesString, "gameContentType=games")[0]
		fmt.Println("Total PS pages string = " + totalPagesString)
		re := regexp.MustCompile("[0-9]+")
		totalPagesString = re.FindAllString(totalPagesString, 1)[0]
		var err error
		count, err = strconv.Atoi(totalPagesString)
		if err != nil {
			count = 1
		}
	} else {
		count = 1
	}
	return count
}

func PsSaveGamesToRepository(games []GameGeneral) {
	gamesData := filetool.CreateFile(activeRepositoryPs)
	filetool.AppendToFile(gamesData, "{\n\"games\": [\n")
	for index, game := range games {
		gameString, _ := json.Marshal(game)
		gameString = bytes.Replace(gameString, []byte("\\u003c"), []byte("<"), -1)
		gameString = bytes.Replace(gameString, []byte("\\u003e"), []byte(">"), -1)
		gameString = bytes.Replace(gameString, []byte("\\u0026"), []byte("&"), -1)
		filetool.AppendToFile(gamesData, string(gameString))
		if index < (len(games) - 1) {
			filetool.AppendToFile(gamesData, ",\n")
		} else {
			filetool.AppendToFile(gamesData, "\n]\n}")
		}
	}
	filetool.CloseFile(gamesData)
	fmt.Println("PsSaveGamesToRepository.Completed")
}

func PsLoadGamesFromRepo(loadArgument string, discountBorder int64, discountRange int64) (games []GameGeneral, isRepoFileExist bool) {
	isRepoFileExist = false
	gamesListRaw, loadError := ioutil.ReadFile(activeRepositoryPs)
	if loadError != nil {
		return games, isRepoFileExist
	} else {
		isRepoFileExist = true
	}

	jsonObjects := gjson.Get(string(gamesListRaw), "games")

	for _, object := range jsonObjects.Array() {
		var tmpGame GameGeneral
		json.Unmarshal([]byte(object.String()), &tmpGame)
		switch loadArgument {
		case "all":
			games = append(games, tmpGame)
		case "discounted":
			if tmpGame.Discount <= discountBorder+discountRange && tmpGame.Discount >= discountBorder+discountRange {
				games = append(games, tmpGame)
			}
		case "allDiscounted":
			if tmpGame.Discount > 0 {
				games = append(games, tmpGame)
			}
		case "notYetPublished":
			if !tmpGame.AlreadyPublished {
				games = append(games, tmpGame)
			}
		}
	}

	return games, isRepoFileExist
}

func randomPs(min, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return rand.Intn(max-min) + min
}
