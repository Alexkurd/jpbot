package main

func getRank(rankId int) string {
	return "Rank"
}

func rankUp() {

}

func readRanks() {
	//Get table userid:messageCount to memory
	//every 50 messages or on rankUp save progress
}

/*func checkAndNotifyRank(bot *tgbotapi.BotAPI, update tgbotapi.Update, userID int, messageCount int, ranks map[string]string) {
	// Example: Check if the message count reaches a certain threshold for rank upgrade
	// This logic can be more complex based on how you want to calculate ranks
	var newRank string
	if messageCount == 10 {
		newRank = ranks["10"]
	} else if messageCount == 50 {
		newRank = ranks["50"]
	} else if messageCount == 100 {
		newRank = ranks["100"]
	}

	if newRank != "" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Congratulations, you've been promoted to "+newRank+"!")
		bot.Send(msg)
	}
}*/
