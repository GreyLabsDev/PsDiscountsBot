package main

import (
	"fmt"
	"os"

	datasource "psDicountsBot/datasource"
	logger "psDicountsBot/logger"
	taskmanager "psDicountsBot/taskmanager"
	telegramBot "psDicountsBot/telegram"
)

func main() {

	logger.Init()

	fmt.Println("Search core started\n")
	fmt.Println()

	exitChannel := make(chan string)

	datasource.PsSwitchProxy(true)
	datasource.PsSetActiveRepository("ps_discounts.json")

	var psFastStartTasks []taskmanager.SingleTask
	var psPostingStartTasks []taskmanager.SingleTask

	psGameFetchTask := datasource.PsGetUpdateDiscountedGamesTask()
	psInitializationTask := datasource.PsGetInitForPublicationTask()
	botPostingBundlePsTask := telegramBot.GetPsPostGameBundleTask(3)

	psFastStartTasks = append(psFastStartTasks,
		psGameFetchTask,
		psInitializationTask,
	)

	psPostingStartTasks = append(psPostingStartTasks,
		botPostingBundlePsTask,
	)

	discountsRefreshEndMessage := "discounts.update"

	var playStationUpdateDiscountsTask taskmanager.PeriodicTask = func() {
		taskmanager.CompleteTaskQueue(psFastStartTasks, discountsRefreshEndMessage, exitChannel)
	}

	var playStationBundlePostingTask taskmanager.PeriodicTask = func() {
		taskmanager.CompleteTaskQueue(psPostingStartTasks, discountsRefreshEndMessage, exitChannel)
	}

	go taskmanager.DoPeriodicTaskAtTime(
		"6:00",
		exitChannel,
		playStationUpdateDiscountsTask,
	)

	go taskmanager.DoPeriodicTaskAtTime(
		"12:00",
		exitChannel,
		playStationBundlePostingTask,
	)

	go taskmanager.DoPeriodicTaskAtTime(
		"14:00",
		exitChannel,
		playStationBundlePostingTask,
	)

	for {
		select {
		case msg := <-exitChannel:
			{
				switch msg {
				case "end":
					fmt.Println("All tasks finished")
					os.Exit(0)
				case "error.network":
					fmt.Println("Network error, mb can`t create request")
				case "discounts.update":
					fmt.Println("All discounts is updated. PlayStation.")
					logger.Write("All discounts is updated. PlayStation.")
				default:
					logger.Write(msg)
				}
			}
		}
	}
}
