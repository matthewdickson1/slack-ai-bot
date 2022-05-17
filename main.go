package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Krognol/go-wolfram"
	"github.com/joho/godotenv"
	"github.com/shomali11/slacker"
	"github.com/tidwall/gjson"
	witai "github.com/wit-ai/wit-go/v2"
	"log"
	"os"
)

var wolframClient *wolfram.Client

func printCommandEvents(analyticsChannel <-chan *slacker.CommandEvent) {
	for event := range analyticsChannel {
		fmt.Println("Command Events")
		fmt.Println(event.Timestamp)
		fmt.Println(event.Command)
		fmt.Println(event.Parameters)
		fmt.Println(event.Event)
		fmt.Println()
	}

}
func main() {
	godotenv.Load(".env")

	bot := slacker.NewClient(os.Getenv("SLACK_BOT_TOKEN"), os.Getenv("SLACK_APP_TOKEN"))
	client := witai.NewClient(os.Getenv("WIT_AI_TOKEN"))
	wolframClient := &wolfram.Client{AppID: os.Getenv("WOLFRAM_APP_ID")}

	go printCommandEvents(bot.CommandEvents())

	bot.Command("Query - <message>", &slacker.CommandDefinition{
		Description: "Send any question to Wolfram",
		Example:     "Who is the president of the USA",
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			query := request.Param("message")
			fmt.Println(query)
			msg, _ := client.Parse(&witai.MessageRequest{
				Query: query,
			})
			//4 spaces
			data, _ := json.MarshalIndent(msg, "", "    ")
			rough := string(data[:])
			value := gjson.Get(rough, "entities.wit$wolfram_search_query:wolfram_search_query.0.value")
			answer := value.String()
			res, err := wolframClient.GetSpokentAnswerQuery(answer, wolfram.Metric, 1000)
			if err != nil {
				fmt.Println("There is an error with your query")
			}

			fmt.Println(res)
			response.Reply(res)
		},
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := bot.Listen(ctx)

	if err != nil {
		log.Fatal(err)
	}
}
