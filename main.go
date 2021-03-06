package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Status struct {
	Login       string
	BlameStatus BlameStatus
}

type BlameStatus struct {
	Received int
	Sent     int
}

type Params struct {
	apiKey         string
	statusFilePath string
}

var BLAME = []string{
	"иди нахуй", "иди на хуй", "инна", "идинахуй", "иннах", "иди нах", "идинах",
	"нахуй иди", "на хуй иди", "нахуйиди", "нахиди",
	"idi nahuj", "idinah", "idinahuj",
	"nahuj idi", "nahidi", "nahujidi",
	"пошел нахуй", "пошел нах", "пошел на хуй", "пошелнахуй",
	"нах пошел", "нахпошел", "на хуй пошел", "нахупошел",
	"пошёл нахуй", "пошёл нах", "пошёл на хуй", "пошёлнахуй",
	"нах пошёл", "нахпошёл", "на хуй пошёл", "нахупошёл"}

const STATUS_CMD = "/status"
const HELP_CMD = "/help"

func main() {
	params := parseArgs()
	bot, err := tgbotapi.NewBotAPI(params.apiKey)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	login2StatusMap := readStatusFromFile(params.statusFilePath)
	log.Printf("statuses initialized: %v", login2StatusMap)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		if isStatusMsg(update) {
			processStatus(update, login2StatusMap, bot)
			continue
		}

		if isBlameMsg(update) {
			login := getLogin(update)
			replyToLogin := getReplyToLogin(update)
			if login == replyToLogin {
				receivedLogin, receivedRate := updateReceived(update, params.statusFilePath, login2StatusMap)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID,
					fmt.Sprintf("%s, сам себя послал нахуй. Адекватность обновлена: %s",
						receivedLogin, receivedRate))
				msg.ReplyToMessageID = update.Message.MessageID
				bot.Send(msg)
				continue
			}
			receivedLogin, receivedRate := updateReceived(update, params.statusFilePath, login2StatusMap)
			sentLogin, sentRate := updateSent(update, params.statusFilePath, login2StatusMap)
			login2StatusMap = readStatusFromFile(params.statusFilePath)

			msg := tgbotapi.NewMessage(update.Message.Chat.ID,
				fmt.Sprintf("%s, послал нахуй %s. Адекватность обновлена: %s: %s; %s: %s",
					sentLogin, receivedLogin, receivedLogin, receivedRate, sentLogin, sentRate))
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
			continue
		}

		if isHelpMsg(update) {
			userFriendlyBlameList, _ := json.MarshalIndent(BLAME, "", "  ")
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Бот измеряет адекватность, беря в расчет сколько раз человек послал нахуй кого-то и сколько раз был послан сам. Единственная доступная команда - /status с помощью которой можно посмотреть свое текущее значение адекватности, или кого-то еще, если написать /status @username. \n\n Послание нахуй считается если сообщение помечено как \"ответ\" и содержит одно из %v. \n\n Одно сообщение влияет одновременно на показатели адекватности посылающего и посланного.", string(userFriendlyBlameList)))
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
			continue
		}

	}
}

func isReply(update tgbotapi.Update) bool {
	return update.Message.ReplyToMessage != nil
}

func isBlameMsg(update tgbotapi.Update) bool {
	if !isReply(update) {
		return false
	}
	for _, blameMsg := range BLAME {
		if strings.Contains(strings.ToLower(update.Message.Text), blameMsg) {
			return true
		}
	}

	return false
}

func isStatusMsg(update tgbotapi.Update) bool {
	return strings.HasPrefix(update.Message.Text, STATUS_CMD)
}

func isHelpMsg(update tgbotapi.Update) bool {
	return strings.HasPrefix(update.Message.Text, HELP_CMD)
}

func processStatus(update tgbotapi.Update, login2StatusMap map[string]Status, bot *tgbotapi.BotAPI) {
	login := getLogin(update)

	txt := strings.ReplaceAll(update.Message.Text, "/status", "")
	txt = strings.TrimSpace(txt)
	statusAbout := strings.ReplaceAll(txt, "@", "")
	if len(statusAbout) < 1 {
		statusAbout = login
	}
	status, ok := login2StatusMap[statusAbout]
	if ok {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			fmt.Sprintf("%s, твоя адекватность %s",
				statusAbout, prettyPrintStatus(status)))
		msg.ReplyToMessageID = update.Message.MessageID
		bot.Send(msg)
	} else {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			fmt.Sprintf("%s, твоя адекватность мне неизвестна еще, попробуй послать кого-нибудь нахуй",
				statusAbout))
		msg.ReplyToMessageID = update.Message.MessageID
		bot.Send(msg)
	}
}

func updateSent(update tgbotapi.Update, filePath string, login2StatusMap map[string]Status) (string, string) {
	login := getLogin(update)

	status, ok := login2StatusMap[login]
	if ok {
		status.BlameStatus.Sent = status.BlameStatus.Sent + 1
	} else {
		status = Status{
			Login: login,
			BlameStatus: BlameStatus{
				Sent:     1,
				Received: 0,
			},
		}
	}
	login2StatusMap[login] = status
	writeStatusToFile(filePath, login2StatusMap)
	return login, prettyPrintStatus(status)
}

func getReplyToLogin(update tgbotapi.Update) string {
	replyToLogin := update.Message.ReplyToMessage.From.UserName
	if len(replyToLogin) < 1 {
		replyToLogin = update.Message.ReplyToMessage.From.FirstName
		if len(update.Message.ReplyToMessage.From.LastName) > 0 {
			replyToLogin = replyToLogin + "_" + update.Message.ReplyToMessage.From.LastName
		}
	}
	return replyToLogin
}

func getLogin(update tgbotapi.Update) string {
	login := update.Message.From.UserName
	if len(login) < 1 {
		login = update.Message.From.FirstName
		if len(update.Message.From.LastName) > 0 {
			login = login + "_" + update.Message.From.LastName
		}
	}
	return login
}

func updateReceived(update tgbotapi.Update, filePath string, login2StatusMap map[string]Status) (string, string) {
	replyToLogin := getReplyToLogin(update)

	status, ok := login2StatusMap[replyToLogin]
	if ok {
		status.BlameStatus.Received = status.BlameStatus.Received + 1
	} else {
		status = Status{
			Login: replyToLogin,
			BlameStatus: BlameStatus{
				Sent:     0,
				Received: 1,
			},
		}
	}
	login2StatusMap[replyToLogin] = status
	writeStatusToFile(filePath, login2StatusMap)
	return replyToLogin, prettyPrintStatus(status)
}

func prettyPrintStatus(status Status) string {
	var rawResult = 100.0
	var sent = 0.1
	var received = 0.1
	if status.BlameStatus.Received != 0 {
		received = float64(status.BlameStatus.Received)
	}
	if status.BlameStatus.Sent != 0 {
		sent = float64(status.BlameStatus.Sent)
	}
	if received > sent {
		rawResult = sent / received * 100.0
	} else {
		rawResult = received / sent * 100.0
	}
	return fmt.Sprintf("%.1f%%", rawResult)
}

func readStatusFromFile(filePath string) map[string]Status {
	login2StatusMap := make(map[string]Status)

	rawStatuses := readfile(filePath)
	log.Printf("Read file: %s", rawStatuses)
	for _, line := range rawStatuses {
		if len(line) < 1 {
			continue
		}
		var status Status
		err := json.Unmarshal([]byte(line), &status)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Unmarshelled: %v", status)
		login2StatusMap[status.Login] = status
	}

	return login2StatusMap
}

func writeStatusToFile(filePath string, login2StatusMap map[string]Status) {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}
	datawriter := bufio.NewWriter(file)

	for _, status := range login2StatusMap {
		jsonStatus, err := json.Marshal(status)
		if err != nil {
			log.Fatal(err)
		}
		_, _ = datawriter.WriteString(string(jsonStatus) + "\n")
	}

	datawriter.Flush()
	file.Close()
}

func parseArgs() Params {
	if len(os.Args) < 3 {
		log.Fatal("two args must be provided: path to the status file; " +
			"telegram Bot API Key")
	}

	var params Params
	params.statusFilePath = os.Args[1]
	params.apiKey = os.Args[2]

	return params
}

func writefile(path string, data []byte) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
	datawriter := bufio.NewWriter(file)
	_, _ = datawriter.WriteString(string(data))
	datawriter.Flush()
	file.Close()
}

func readfile(path string) []string {
	buff := make([]string, 0)
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		buff = append(buff, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return buff
}
