package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/kyokomi/emoji"
	"github.com/nlopes/slack"
	"github.com/tarm/serial"
)

var state State
var status Status
var port *serial.Port
var scanner *bufio.Scanner
var client *slack.Client
var cTrinket chan bool
var cSlack chan bool
var cStatus chan bool
var cScanner chan string
var cQuit chan bool

// State represents device connection statuses.
type State struct {
	trinket bool
	slack   bool
}

// Status represents a users slack statuses.
type Status struct {
	emoji string // "[:emoji:]"
	text  string // "[status text]"
}

// Open Port
//
// Helper function to open serial connection to provided device.
func openPort(path string) (*serial.Port, error) {
	c := new(serial.Config)
	c.Name = path
	c.Baud = 115200
	c.Size = 8
	c.Parity = 'N'
	c.StopBits = 1

	return serial.OpenPort(c)
}

// Run Scanner
//
// Function to be used as a goroutine and communicate any text received
// over the cScanner channel.  Listens for kill signal on quit channel so it
// can be restarted in case of a loss of connection to the trinket.
//
// We need to ignore certain strings as we can't distinguish between what
// has been sent to and what has been received from the trinket.  Scanner will
// only return one message to the channel
func runScanner() {
	for scanner.Scan() {
		select {
		case <-cQuit:
			return
		default:
			cScanner <- scanner.Text()
		}
	}
}

// Connect to Trinket
//
// 1. Grab the contents of /dev
// 2. Loop through each file looking for *tty.usbmodem*
// 3. For each possible match establish a serial connection and say 'hey'
// 4. If we get a response back of 'go away' we've found our guy, keep port and scanner, then notify channel
// 5. If no 'go away' response after a half second close the serial port and stop the scanner, try again in 5 seconds
// * runScanner() goroutine will be closed out when this function returns
func findTrinket() {
	fmt.Printf("locating trinket")

	for {
		// grab the contents of /dev
		contents, err := ioutil.ReadDir("/dev")
		if err != nil {
			fmt.Println(err)
			time.Sleep(time.Millisecond * 5000)
			continue
		}

		// loop through contents of /dev
		for _, f := range contents {
			// look for what is mostly likely the trinket device
			if strings.Contains(f.Name(), "tty.usbmodem") {
				// initialize port and scanner
				port, _ = openPort("/dev/" + f.Name())
				scanner = bufio.NewScanner(port)

				// start listening for responses
				go runScanner()

				// send our test message
				port.Write([]byte("hey\r"))

				// pull off the first buffered item
				fmt.Println("msg sent: " + <-cScanner)

				// listen 500ms for a response
				select {
				case received := <-cScanner:
					fmt.Println("msg received: " + received)
					if received != "go away" {
						port.Close()
						cQuit <- true
					} else {
						fmt.Printf("\nfound trinket: %s\n", string(f.Name()))
						cTrinket <- true
						return
					}
				case <-time.After(time.Millisecond * 500):
					fmt.Println("timed out")
					port.Close()
					cQuit <- true
				}
			}
		}

		// try again in 5 seconds
		fmt.Printf(".")
		time.Sleep(time.Millisecond * 5000)
	}
}

func monitorTrinket() {
	go runScanner()
	for {
		port.Write([]byte("ping\r"))
		select {
		case received := <-cScanner:
			if received == "pong" {
				fmt.Println("trinket heartbeat received")
			} else {
				fmt.Println("trinket heartbeat missed")
				cTrinket <- false
			}
		case <-time.After(time.Millisecond * 500):
			fmt.Println("trinket heartbeat missed")
			cTrinket <- false
		}
		time.Sleep(time.Millisecond * 10000)
	}
}

// Get Current Slack Status
//
// 1. Pulls the current user status (text and emoji) from slack.
// 2. Updates the global status.
// 3. Notifies the cStatus channel so additional actions can be hooked.
// 2. If unable to get a status, waits 10 seconds and tries again.
func getCurrentSlackStatus() {
	var user *slack.User
	var err error
	for {
		user, err = client.GetUserInfo(slackUserID)
		if err == nil {
			emoji.Println("received slack status: " + user.Profile.StatusEmoji + " " + user.Profile.StatusText)
			status.emoji = user.Profile.StatusEmoji
			status.text = user.Profile.StatusText
			cStatus <- true
			return
		}

		fmt.Printf("error getting user status: %s\n", err)
		time.Sleep(time.Millisecond * 10000)
	}
}

func runSlack(c chan<- string) {
	// 1. wait until trinket is connected
	//
	// setup our slack client
	client = slack.New(
		slackLegacyToken,
		slack.OptionDebug(false),
		slack.OptionLog(log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)),
	)

	getCurrentSlackStatus()

	// start new rtm client connection to monitor for status changes
	rtm := client.NewRTM()
	go rtm.ManageConnection()

	// loop through events and watch for any changes to status
	for msg := range rtm.IncomingEvents {
		fmt.Print("Event Received: ")
		switch ev := msg.Data.(type) {
		case *slack.UserChangeEvent:
			fmt.Printf("%T\n", ev)
			//fmt.Printf("EventVals: %+v\n", ev)
			if ev.User.ID == slackUserID {
				//setLightStatus(strings.ToLower(ev.User.Profile.StatusText), ev.User.Profile.StatusEmoji)
			}
		default:
			fmt.Printf("%T\n", ev)
		}
	}

	//c <- "connected"
}

func main() {
	// channels for communicating changes to states or status
	cTrinket = make(chan bool)      // change in trinket state
	cSlack = make(chan bool)        // change in slack state
	cStatus = make(chan bool)       // change in slack status
	cScanner = make(chan string, 2) // message received from serial connection, buffer size of two to capture msg, response
	cQuit = make(chan bool)         // for killing the scanner

	// kick off our goroutines
	go findTrinket() // establishes serial port with trinket
	//go runSlack(cSlack) // establishes rtm api connection & grabs changes in state

	// startup our event listeners
	for {
		select {
		case trinketConnected := <-cTrinket:
			// trinket connection has changed
			if trinketConnected {
				monitorTrinket()
			} else {
				findTrinket()
			}
		case <-cSlack:
			// slack connection has changed
			fmt.Println("received on cSlack channel")
		case <-cStatus:
			// status has changed
			fmt.Println("received on cStatus channel")
			// set the color of the light
			//setLightStatus(strings.ToLower(status.text), status.emoji)
		}
	}
}
