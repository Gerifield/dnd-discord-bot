package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

func main() {
	botToken := flag.String("botToken", "", "Token for the bot to auth with")
	flag.Parse()

	discord, err := discordgo.New("Bot " + *botToken)
	if err != nil {
		log.Fatalln(err)
	}

	discord.AddHandler(messageHandle)
	discord.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages)

	// Open a websocket connection to Discord and begin listening.
	err = discord.Open()
	if err != nil {
		log.Fatalln("error opening connection", err)
	}

	// Wait here until CTRL-C or other term signal is received.
	log.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	_ = discord.Close()
}


// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageHandle(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// If the message is "ping" reply with "Pong!"
	if strings.HasPrefix(m.Content, "/roll ") {
		dice := strings.ToLower(strings.TrimPrefix(m.Content,"/roll "))
		diceParams := strings.Split(dice, "d")
		if len(diceParams) < 2 {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Invalid dice (use xdy format)")
			return
		}

		diceNum, err := strconv.ParseInt(diceParams[0], 10, 32)
		if err != nil {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Invalid dice (use xdy format)")
			return
		}

		diceSize, err := strconv.ParseInt(diceParams[1], 10, 32)
		if err != nil {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Invalid dice (use xdy format)")
			return
		}

		if diceNum > 100 {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Use maximum of 100 dice")
			return
		}

		if diceSize > 1000 {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Use maximum of 1000 sided dice")
			return
		}

		res, resSum, err := dropDice(diceNum, diceSize)
		if err != nil {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Whoops, I lost the dice")
			log.Println("random generator failure", err)
			return
		}

		log.Printf("Drop with %d, %d dice for %s and the result is %v, sum %d", diceNum, diceSize, m.Author.Username, res, resSum)
		_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s: %d drop(s) with a %d sided dice are [%s] and sum is %d", m.Author.Mention(), diceNum, diceSize, strings.Join(res, ", "), resSum))
	}
}

func dropDice(diceNum, diceSize int64) ([]string, int64, error) {
	var result []string
	var resultSum int64

	for i:=0; i < int(diceNum); i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(diceSize))
		if err != nil {
			return result, resultSum, err
		}
		result = append(result, n.String())
		resultSum += n.Int64()
	}
	return result, resultSum, nil
}