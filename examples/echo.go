package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	line "github.com/stanaka/line-bot-api"
)

func main() {
	http.HandleFunc("/callback", handler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := fmt.Sprintf(":%s", port)
	http.ListenAndServe(addr, nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	api := line.New(
		os.Getenv("LINE_CHANNEL_ID"),
		os.Getenv("LINE_CHANNEL_SECRET"),
		os.Getenv("LINE_MID"),
	)
	err := api.SetProxy(os.Getenv("PROXY_URL"))
	if err != nil {
		log.Println(err)
	}
	msg, err := api.DecodeMessage(r.Body)
	if err != nil {
		log.Println(err)
	}
	for _, result := range msg.Results {
		from := result.Content.From
		text := result.Content.Text
		err := api.SendMessage([]string{from}, text)
		if err != nil {
			log.Println(err)
		}
	}
	fmt.Fprintf(w, "OK")
}
