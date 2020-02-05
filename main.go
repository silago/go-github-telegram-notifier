package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	var telegramChatID = os.Getenv("TELEGRAM_CHAT_ID")
	var telegramBotAPIKey = os.Getenv("TELEGRAM_BOT_API_KEY")
	var port = os.Getenv("PORT")

	log.Println("ready")
	var notifier = &Notifier{
		ChatId: telegramChatID,
		ApiKey: telegramBotAPIKey,
	}
	http.HandleFunc("/", notifier.Handler)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

type PushPayload struct {
	Repository BaseEntity
	Pusher     BaseEntity
	Commits    []Commit
}

type BaseEntity struct {
	Name string
}

type Commit struct {
	Committer BaseEntity
	Message   string
}

type Notifier struct {
	ChatId string
	ApiKey string
}

const push_event_type = "push"
const pull_event_type = "pull_request"
const pull_review_type = "pull_request_review"

func (n *Notifier) Handler(writer http.ResponseWriter, r *http.Request) {
	var response string = ""
	var process_error error
	var processor func(*http.Request) (string, error)

	if _, err := writer.Write([]byte(response)); err != nil {
		log.Println(err.Error())
	}
	//var request_type string = r.Header.Get("HTTP_X_GITHUB_EVENT")
	var request_type string = os.Getenv("Http_X_Github_Event")
	for name, values := range r.Header {
	    // Loop over all values for the name.
	    for _, value := range values {
		fmt.Println(name, value)
	    }
}

	switch request_type {
	case push_event_type:
		{
			processor = n.ProcessPush
		}
	default:
		{
			log.Printf("unsupported request type %s", request_type)
			return
		}
	}

	if response, process_error = processor(r); process_error != nil {
		log.Println("[error]" + process_error.Error())
	} else {
		n.SendMessage(response)
	}
}

func (n *Notifier) SendMessage(msg string) {
	msg = strings.ReplaceAll(msg, "\n", "%0A")
	msg = strings.ReplaceAll(msg, "\r", "")
	var url string = fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?chat_id=%s&disable_web_page_preview=1&text=%s", n.ApiKey, n.ChatId, msg)
	if response, err := http.Get(url); err != nil {
		log.Println(err.Error())
	} else if body, err := ioutil.ReadAll(response.Body); err != nil {
		log.Println(err.Error())
	} else {
		log.Println(string(body))
	}

	//msg = fmt.re
	//$text = fmt.Re("\n", "%0A", $text);
	//$method_url = 'https://api.telegram.org/bot' . TELEGRAM_BOT_API_KEY . '/sendMessage';

	//$url = $method_url . '?chat_id=' . TELEGRAM_CHAT_ID . '&disable_web_page_preview=1&text=' . $text;

	//$response = @file_get_contents($url);
	//error_log(var_export($response, true));
	//error_log(" SEND MESSAGE" . $text);

}

func (*Notifier) ProcessPush(r *http.Request) (result string, err error) {
	var data PushPayload
	//var result string = ""
	//var err error

	//defer func() {
	//	return result, err
	//}()

	if err = r.ParseForm(); err != nil {
		return result, err
	}

	var payload = r.FormValue("payload")
	if payload == "" {
		err = errors.New("empty payload data")
		return result, err
	}

	if err = json.Unmarshal([]byte(payload), &data); err != nil {
		return result, err
	}

	var commits_message string = ""
	for _, commit := range data.Commits {
		commits_message += fmt.Sprintf("\r\nCommiter: %s; Text: %s ", commit.Committer.Name, commit.Message)
	}

	result = fmt.Sprintf("New push in %s made by %s %s ", data.Repository.Name, data.Pusher.Name, commits_message)
	return result, err
}
