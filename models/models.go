package models

import (
	"strconv"
	"strings"
)

type GameGeneral struct {
	GlobalID         string      `json:"GlobalID"` // main filed (!) ST125 - steam, 125`th in steam games list, GG - gog, GB - gabestore
	Name             string      `json:"Name"`     // main filed (!)
	HeaderImageURL   string      `json:"HeaderImageURL"`
	IsFree           bool        `json:"IsFree"`
	Discount         int64       `json:"Discount"` // main filed (!)
	Price            int64       `json:"Price"`    // main filed (!)
	Link             string      `json:"Link"`
	Source           string      `json:"Source"` // main filed (!)
	Metacritic       int64       `json:"Metacritic"`
	SteamID          int64       `json:"SteamID"` // main filed (!)
	IsSteamBundle    bool        `json:"IsSteamBundle"`
	SteamBundle      SteamBundle `json:"SteamBundle"`
	AlreadyPublished bool        `json:"AlreadyPublished"` // main filed (!)
}

type SteamBundle []int

type Applist struct {
	Applist Apps `json:"applist"`
}

type Apps struct {
	Apps []App `json:"apps"`
}

type App struct {
	AppID int    `json:"appid"`
	Name  string `json:"name"`
}

func MergeGameLists(oldList []GameGeneral, newList []GameGeneral) (mergedList []GameGeneral) {
	mergedList = oldList

	lastGlobalIDString := oldList[(len(oldList) - 1)].GlobalID
	lastGlobalIDString = strings.Split(lastGlobalIDString, "_")[1]

	lastGlobalIDNumber, _ := strconv.Atoi(lastGlobalIDString)
	lastGlobalIDNumber++

	for _, game := range newList {
		if !ContainsGameGeneral(oldList, game) {
			game.GlobalID = strings.Split(game.GlobalID, "_")[0] + "_" + strconv.Itoa(lastGlobalIDNumber)
			mergedList = append(mergedList, game)
			lastGlobalIDNumber++
		}
	}

	for index, oldGame := range mergedList {
		if !ContainsGameGeneral(newList, oldGame) {
			copy(mergedList[index:], mergedList[index+1:])
			mergedList = mergedList[:len(mergedList)-1]
		}
	}

	return mergedList
}

func ContainsGameGeneral(listForCheck []GameGeneral, gameForCheck GameGeneral) bool {
	for _, game := range listForCheck {
		if game.Link == gameForCheck.Link {
			return true
		}
	}
	return false
}

func GetIndexByGlobalId(games []GameGeneral, globalId string) (outIndex int) {
	outIndex = -1
	for index, game := range games {
		if game.GlobalID == globalId {
			return index
		}
	}
	return outIndex
}
