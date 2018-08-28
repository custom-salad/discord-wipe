package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	l *log.Logger

	messagesParsed  int
	messagesDeleted int

	lowMSWaitFlag  int
	highMSWaitFlag int

	messagesToRetrieveFlag int

	errNoMessagesFromMe = fmt.Errorf("none of my messages found")
)

const (
	error500RetryAttempts     = 5
	error500RetryDelaySeconds = 10

	minimumMSWait = 700
)

func main() {
	var (
		serverIDFlag   string
		exemptionsFlag string
		onlyWipeFlag   string
	)

	flag.StringVar(&serverIDFlag, "server", "", "server ID to wiperoni")
	flag.StringVar(&exemptionsFlag, "exempt", "", "comma-delineated channels to ignore")
	flag.StringVar(&onlyWipeFlag, "wipechannels", "", "comma-delineated list of channels to only wipe")
	flag.IntVar(&lowMSWaitFlag, "waitmin", 1100, fmt.Sprintf("set a minimum randomized wait time. minimum is %dms", minimumMSWait))
	flag.IntVar(&highMSWaitFlag, "waitmax", 1900, "set a maximum randomized wait time")
	flag.IntVar(&messagesToRetrieveFlag, "retrieve", 50, "# of messages to retrieve per loop. 25 <= retrieve <= 100")

	flag.Parse()
	if len(serverIDFlag) == 0 || lowMSWaitFlag < minimumMSWait ||
		highMSWaitFlag < minimumMSWait || lowMSWaitFlag > highMSWaitFlag ||
		messagesToRetrieveFlag < 25 || messagesToRetrieveFlag > 100 {

		flag.Usage()
		os.Exit(2)
	}
	readToken, err := ioutil.ReadFile("token.txt")
	if err != nil {
		l.Println("error loading token.txt, does this file exist?")
		l.Fatal(err)
	}

	d, err := discordgo.New(fmt.Sprintf("%s", strings.TrimSpace(string(readToken))))
	if err != nil {
		l.Fatal(err)
	}
	defer d.Close()
	if err = d.Open(); err != nil {
		l.Fatal(err)
	}
	fmt.Println(ascii)
	channels, exemptions, toWipe := parseChannels(d, serverIDFlag, strings.Split(exemptionsFlag, ","), strings.Split(onlyWipeFlag, ","))
	g, err := d.Guild(serverIDFlag)
	if err != nil {
		l.Fatal(err)
	}
	fmt.Printf("Performing a wipe on the server %s\n", g.Name)
	fmt.Printf("Wiping these channels:\n%s\n\n\n", strings.Join(toWipe, "\n"))
	fmt.Printf("Exempting these channels:\n%s\n\nOk? (y/n): ", strings.Join(exemptions, "\n"))
	scan := bufio.NewScanner(os.Stdin)
	scan.Scan()
	if scan.Text() != "y" && scan.Text() != "yes" {
		l.Println("Cancelled job")
		os.Exit(0)
	}
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for range ticker.C {
			l.Printf("Deleted %d / %d msgs\n", messagesDeleted, messagesParsed)
		}
	}()
	wipe(d, channels)
	l.Printf("Done!\n\nWiped %d / %d msgs\n", messagesDeleted, messagesParsed)
}

func wipe(d *discordgo.Session, channels []*discordgo.Channel) {
	for _, ch := range channels {
		l.Printf("Performing wipe on %s\n", ch.Name)
		var (
			lastMsgID   = ch.LastMessageID
			toDelete    []string
			delayDelete string
			err         error
		)
		for {
			errCounter := 0

		recheck:
			if toDelete, lastMsgID, delayDelete, err = wipeHelper(d, ch, lastMsgID); err != nil {
				if strings.Contains(err.Error(), "500") {
					errCounter++
					if errCounter > error500RetryAttempts {
						l.Fatal(err)
					}
					fmt.Println("Waiting on 500")
					time.Sleep(error500RetryDelaySeconds * time.Second)
					goto recheck
				}
			}
			if len(toDelete) == 0 && err != errNoMessagesFromMe {
				break
			}
			err = wipeDeleteMsgs(d, ch, toDelete)
			if len(delayDelete) > 0 {
				// To avoid dereferencing a null pointer, wait 1.5 seconds
				// to allow the message search to continue past this point
				go func() {
					time.Sleep(1500)
					d.ChannelMessageDelete(ch.ID, delayDelete)
				}()
			}
			if err != nil {
				l.Fatal(err)
			}
		}
	}
}

func wipeDeleteMsgs(d *discordgo.Session, ch *discordgo.Channel, msgs []string) error {
	for _, msg := range msgs {
		retryTimeout := 0
		for {
			err := d.ChannelMessageDelete(ch.ID, msg)
			if err != nil {
				if strings.Contains(err.Error(), "500") {
					retryTimeout++
					if retryTimeout >= error500RetryAttempts {
						return err
					}
					fmt.Println("Waiting on 500")
					time.Sleep(error500RetryDelaySeconds * time.Second)
					continue
				}
				return err
			}
			l.Printf("Deleted %s from #%s\n", msg, ch.Name)
			break
		}
		messagesDeleted++
		time.Sleep(time.Duration(randomDuration(lowMSWaitFlag, highMSWaitFlag)) * time.Millisecond)
	}
	return nil
}

func randomDuration(a, b int) int {
	return rand.Intn(b-a) + a
}

func wipeHelper(d *discordgo.Session, ch *discordgo.Channel, before string) (toDelete []string, newBefore, delayDelete string, err error) {
	msgs, err := d.ChannelMessages(ch.ID, messagesToRetrieveFlag, before, "", "")
	if err != nil {
		return nil, "", "", err
	}
	if len(msgs) == 0 {
		return make([]string, 0), "", "", fmt.Errorf("no new messages")
	}
	for idx, msg := range msgs {
		if idx == len(msgs)-1 {
			newBefore = msg.ID
			if msg.Author.ID == d.State.User.ID {
				delayDelete = msg.ID
			}
			break
		}
		messagesParsed++
		if msg.Author.ID == d.State.User.ID {
			toDelete = append(toDelete, msg.ID)
		}
	}
	if len(toDelete) == 0 {
		return nil, newBefore, "", errNoMessagesFromMe
	}
	return
}

func parseChannels(d *discordgo.Session, serverID string, exemptIDs, onlyWipeIDs []string) (ret []*discordgo.Channel, exemptions, toWipe []string) {
	guild, err := d.Guild(serverID)
	if err != nil {
		l.Fatal(err)
	}
	for _, ch := range guild.Channels {
		if ch.Type != discordgo.ChannelTypeGuildText {
			continue
		}
		if _, err := d.Channel(ch.ID); err != nil {
			l.Println(err)
			continue
		}
		if len(onlyWipeIDs) > 0 && len(onlyWipeIDs[0]) > 0 {
			if !contains(ch.ID, onlyWipeIDs) {
				exemptions = append(exemptions, ch.Name)
				continue
			}
		} else {
			if contains(ch.ID, exemptIDs) {
				exemptions = append(exemptions, ch.Name)
				continue
			}
		}
		toWipe = append(toWipe, ch.Name)
		ret = append(ret, ch)
	}
	return
}

func contains(s string, arr []string) bool {
	for _, arrS := range arr {
		if s == arrS {
			return true
		}
	}
	return false
}

func init() {
	l = log.New(os.Stderr, "main: ", log.LstdFlags|log.Lshortfile)
	rand.Seed(time.Now().Unix())
}

const (
	ascii = `
_____ _____  _____  _____ ____  _____  _____   __          _______ _____  ______ 
|  __ \_   _|/ ____|/ ____/ __ \|  __ \|  __ \  \ \        / /_   _|  __ \|  ____|
| |  | || | | (___ | |   | |  | | |__) | |  | |  \ \  /\  / /  | | | |__) | |__   
| |  | || |  \___ \| |   | |  | |  _  /| |  | |   \ \/  \/ /   | | |  ___/|  __|  
| |__| || |_ ____) | |___| |__| | | \ \| |__| |    \  /\  /   _| |_| |    | |____ 
|_____/_____|_____/ \_____\____/|_|  \_\_____/      \/  \/   |_____|_|    |______|

`
)
