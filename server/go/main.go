package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/nlopes/slack"
	"github.com/tarm/serial"
)

const (
	slackUserID      = "UCRP0RS8H"
	slackLegacyToken = "xoxp-420969448000-433782876289-459217447138-f9e84c14e73c035385426d60ea028e57"
)

func initializePort(path string) *serial.Port {
	c := new(serial.Config)
	c.Name = path
	c.Baud = 115200
	c.Size = 8
	c.Parity = 'N'
	c.StopBits = 1

	stream, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}

	return stream
}

func findTrinket() string {
	// grab the contents of /dev
	contents, err := ioutil.ReadDir("/dev")
	if err != nil {
		log.Fatal(err)
	}

	// look for what is mostly likely the trinket device
	for _, f := range contents {
		if strings.Contains(f.Name(), "tty.usbmodem") {
			// initialize stream and scanner
			stream := initializePort("/dev/" + f.Name())
			scanner := bufio.NewScanner(stream)

			// check for 'go away response' to confirm it's our guy
			stream.Write([]byte("hey\r"))
			for scanner.Scan() {
				rcv := scanner.Text()
				if rcv == "go away" {
					fmt.Println("found trinket:", string(f.Name()))
					return "/dev/" + f.Name()
				}
			}
		}
	}

	// unable to find any candidates
	return ""
}

func setLightStatus(stream *serial.Port, statusText string, statusEmoji string) {
	// default solid green
	color := "green"
	mode := "@solid"

	if statusEmoji != "" && statusText == "" {
		// if there is only an emoji status...
		switch statusEmoji {
		case ":middle_finger:":
			color = "red"
			mode = "@pulse"
		case ":red_circle:", ":woman-gesturing-no:", ":man-gesturing-no:", ":male-technologist:", ":female-technologist:":
			color = "red"
			mode = "@solid"
		case ":thinking_face:", ":sleeping:", ":shushing_face:":
			color = "yellow"
			mode = "@solid"
		}
	} else if statusText != "" {
		// else we are just going to use the text status
		switch statusText {
		case "in a meeting", "on a call":
			color = "red"
			mode = "@pulse"
		case "focused":
			color = "red"
			mode = "@solid"
		case "thinking":
			color = "yellow"
			mode = "@solid"
		}
	}

	// set the light
	fmt.Printf("Setting Light: %v-%v\n", mode, color)
	_, err := stream.Write([]byte(color + "\r" + mode + "\r"))
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	// establish our serial stream
	stream := initializePort(findTrinket())

	// establish our slack connection, generate a legacy user token for this connection
	api := slack.New(
		slackLegacyToken,
		slack.OptionDebug(false),
		slack.OptionLog(log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)),
	)

	// grab our current user status and set the light
	user, err := api.GetUserInfo(slackUserID)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	setLightStatus(stream, strings.ToLower(user.Profile.StatusText), user.Profile.StatusEmoji)

	// start new rtm api connection to monitor for status changes
	rtm := api.NewRTM()
	go rtm.ManageConnection()

	// loop through events
	for msg := range rtm.IncomingEvents {
		fmt.Print("Event Received: ")
		switch ev := msg.Data.(type) {
		case *slack.UserChangeEvent:
			fmt.Printf("%T\n", ev)
			//fmt.Printf("EventVals: %+v\n", ev)
			if ev.User.ID == slackUserID {
				setLightStatus(stream, strings.ToLower(ev.User.Profile.StatusText), ev.User.Profile.StatusEmoji)
			}
		default:
			fmt.Printf("%T\n", ev)
		}
	}
}
