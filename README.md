Do Not Disturb
==============

A hardware and software project for those who work in an open office and need a better way than headphones to indicate availability.  The light ties to Slack statuses to tell others how wary they should be to approach you.  There will be some soldering involved, and although this can be done on any OS we'll assume OSX here.

![](assets/IMG_6796.JPG)

| Color  | Behavior | Slack Status Text | Slack Status Emoji | Other |
| :----- | :------- | :---------------- | :----------------- | :---- |
| blue   | pulsing  | - | - | trinket is connected to the go server, but not slack |
| blue   | solid    | - | - | trinket is connected to the go server and slack |
| red    | pulsing  | 'in a meeting', 'on a call' | :middle_finger: | |
| red    | solid    | 'focused', 'busy' | :triangular_flag_on_post:, :red_circle:, :woman-gesturing-no:, :man-gesturing-no:, :male-technologist:, :female-technologist: | |
| green  | pulsing  | - | - | - |
| green  | solid    | __default__ | __default__ | If there is no status or one we don't trigger on the light will default green. |
| yellow | pulsing  | - | - | - |
| yellow | solid    | 'thinking' | :thinking_face:, :sleeping:, :shushing_face: | |

* If there is Status Text then we ignore the Status Emoji, only if there is no Status Text do we set the light based on Status Emoji.

Build
-----

| Trinket M0 Front | Trinket M0 Back |
| :-----------: | :----------: |
| ![front](assets/IMG_0842.JPG) | ![](assets/IMG_7806.JPG) |

### Components

| Item | Quantity | Price |
| ---- | -------- | ----- |
| [Adafruit Trinket M0](https://www.adafruit.com/product/3500) | 1 | ~$9 |
| [Diffused Rectangular 5mm RGB LED (common anode)](https://www.adafruit.com/product/2739) | 1 | ~$6 for 10 |
| [Photoresistor](https://www.adafruit.com/product/161) | 1 | ~$1 |
| [300 Ω Resistors (1206 SMD)](https://www.mouser.com/Passive-Components/Resistors/SMD-Resistors-Chip-Resistors/_/N-7h7yu?P=1z0x8a5Z1z0x6frZ1yzmoty) | 3 | ~$0.30 |
| [10K Ω Resistor (1206 SMD)](https://www.mouser.com/Passive-Components/Resistors/SMD-Resistors-Chip-Resistors/_/N-7h7yu?P=1z0x6frZ1yzmotyZ1yzmno7) | 1 | ~$0.15 |
| [MicroUSB Data Cable](https://www.amazon.com/dp/B0711PVX6Z/ref=cm_sw_em_r_mt_dp_U_NwJdDb4PMCY4R) | 1 | ~$5 |
| [30AWG Stranded-Core Wire (black)](https://www.adafruit.com/product/3164) | 1 inch | ~$5 |

### Wiring

### Tips

We can't control the 'on' led, it will always be green, it will always be on and we really don't want to be giving the wrong idea that you are free to chat.  You've got a lot of options but I recommend getting some liquid electrical tape and just putting a dab of it right on top.  It's non-destructive, blends in with the black solder mask of the board and comes off in one nice chunk if you ever choose to clean it up.

Setup
-----

### Hardware:

1. Build the hardware, see above
2. Upload CircuitPython 4.X firmware to device
   - Download the latest from [here](https://github.com/adafruit/circuitpython/releases<Paste>), it's a large page... search for 'trinket_m0-en_US'
   - Plug in and backup any existing projects on your Trinket M0
   - Double click the Trinket's reset button to enter bootloader mode, the red light will flash and a new drive titled TRINKETBOOT will be mounted
   - Place the downloaded UF2 file into this newly mounted drive, it will automatically be applied and the device will reset
3. Upload CIRCUITPY contents to device
   - Once the firmware is done and the device has reset you'll see a CIRCUITPY device mounted, copy the CIRCUITPY contents of this repository to the newly mounted drive, that's it!

### Slack:

1. Create a legacy Token
2. Get your ID

### Server:

1. Grab dependencies

        go get github.com/tarm/serial
        go get github.com/nlopes/slack
        go get github.com/kyokomi/emoji

2. Add Slack legacy token and user ID to server/go/main.go
3. Compile the application
4. Create a local service so it runs automatically at boot

Develop
-------

You'll want:

- a serial terminal emulator
- your favorite text editor
- a slack api
- ...

### Algoritm

1. connect to trinket
   - look for possible devices
   - message each until we hear a 'go away' response to our 'hey' statement
   - if no success with any candidate retry in 10 seconds
2. get the current status of the slack user
   - if unsuccessful retry in 10 seconds
3. set the trinket light based on the current status
   - if unsuccessful reset trinket connection and try again in, no delay as the connect to trinket func will poll until device comes back online
4. connect to the rtm client and watch for status change events
5. set the trinket light based on the updated status
   - if unsuccessful reset trinket connection and try again in, no delay as the connect to trinket func will poll until device comes back online


slack
1. get current status
  - if unable to connect wait and try again
  - if trinket state is connected then set color
2. listen for status updates
  - if status received and trinket state is connected set trinket color
  - if disconnects start again from 1

trinket
1. look for device
  - list possible devices
  - for each say 'hey'
  - if rcv 'go away' mark as it
  - if none respond, wait and try again
2. listen for color change requests
3. every 5 seconds send ping to device
  - if pong received set state as connected and carry on
  - if no pong received close out stream and start over again at 1

        blue -> not connected to go server
        ??   -> connected to go server but not slack server
        ____ -> connected to both go and slack servers


References
----------

Todo
----

- [ ] clean files, file contents, readme
- [ ] add in retries with progressing timeouts for finding trinket
- [ ] check for / add errors to set light func, attempt to reconnect to trinket
- [ ] situations to handle
      - no trinket
      - no internet
      - combinations of the above
      - trinket removed
      - internet removed
      - combinations of the above
- [ ] add a ping every 5 seconds and if trinket does not pong back then try to re-establish stream, can this be done with channels and goroutines??
- [ ] refactor, simplify
- [ ] clean up output of go application
- [ ] write service wrapper with /var/log/donotdisturb output and log rotation

!! WANT to have one active stream so we connect once?

Colophon
--------

