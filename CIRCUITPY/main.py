#############
#  imports  #
#############

import pulseio
import time
import board
import adafruit_rgbled
import adafruit_dotstar as dotstar
import supervisor
from analogio import AnalogIn

####################
#  configurations  #
####################

# pwm out, order is important here due to limited number of timers on trinket
# most if not all of these pins have access to two timers and grab the first available
# as such you can initialize these in a way that they block others from finding an
# available timer... not sure that's really the case but this order works so no touchy!
pwm3 = pulseio.PWMOut(board.D3, frequency=1000, duty_cycle=65535) # timer 1 (blue)
pwm4 = pulseio.PWMOut(board.D4, frequency=1000, duty_cycle=65535) # timer 1 (green)
pwm2 = pulseio.PWMOut(board.D2, frequency=1000, duty_cycle=65535) # timer 0 (red)
pwm0 = pulseio.PWMOut(board.D0, frequency=1000, duty_cycle=65535) # timer 0 (unused)
led = pulseio.PWMOut(board.D13, frequency=1000, duty_cycle=65536) # timer 0 (on board red)

# analog pin for our photoresistor
light = AnalogIn(board.A0)

#############
#  helpers  #
#############

# convert analog input to voltage
def getVoltage(pin):
    return (pin.value * 3.3) / 65536

# set the led colors
def setColor(color):
    # improve to accept hex ie 0x333333
    brddot[0] = color
    rgbled.color = color

# set luminosity based on ambient light level
def setLuminosity():
    voltageIndex = int(getVoltage(light))

    if voltageIndex == 0:
        luminosity = 10
        pulseTimeout = 0.08
    elif voltageIndex == 1:
        luminosity = 20
        pulseTimeout = 0.04
    elif voltageIndex == 2:
        luminosity = 40
        pulseTimeout = 0.02
    elif voltageIndex == 3:
        luminosity = 80
        pulseTimeout = 0.01
    return luminosity, pulseTimeout


# tuples are immutable, helper to convert to list, replace then convert back
def replace(tup, x, y):
   tup_list = list(tup)
   for element in tup_list:
      if element == x:
         tup_list[tup_list.index(element)] = y
   new_tuple = tuple(tup_list)
   return new_tuple

#####################
#  initializations  #
#####################

# dotstar and rgbled
brddot = dotstar.DotStar(board.APA102_SCK, board.APA102_MOSI, 1, brightness=1)
rgbled = adafruit_rgbled.RGBLED(pwm2, pwm4, pwm3, invert_pwm=True)

# lets default to solid mode, 'pulse' is the alternative
mode = "solid"

# first pass at luminosity, and initialize the oldluminosity var
luminosity, pulseTimeout = setLuminosity()
oldluminosity = luminosity

# default to a yellow to indicate we haven't received anything from the server
color = (0,0,luminosity)
colorName = 'blue'
setColor(color)

##################
#  run the loop  #
##################

while True:
    # grab the current luminosity
    luminosity, pulseTimeout = setLuminosity()

    # if luminosity has changed reset the color!
    if (oldluminosity != luminosity):
        if (colorName == 'red'):
            color = (luminosity,0,0)
        if (colorName == 'green'):
            color = (0,luminosity,0)
        if (colorName == 'yellow'):
            color = (luminosity,luminosity,0)

    # grab the new luminosity value so we can compare it next loop
    oldluminosity = luminosity

    # check for serial input, accepts red, green, yellow, @pulse, @solid
    # setting a color will default to solid, you must follow with a mode
    # as a second input to change to a pulse
    if supervisor.runtime.serial_bytes_available:
        command = input()
        if (command.lower().startswith('@')):
            mode = command[1:]
        elif (command.lower().startswith('red')):
            color = (luminosity,0,0)
            colorName = 'red'
            if (mode == "pulse"):
                mode = "solid"
        elif (command.lower().startswith('green')):
            color = (0,luminosity,0)
            colorName = 'green'
            if (mode == "pulse"):
                mode = "solid"
        elif (command.lower().startswith('yellow')):
            color = (luminosity,luminosity,0)
            colorName = 'yellow'
            if (mode == "pulse"):
                mode = "solid"
        elif (command.lower().startswith('blue')):
            color = (0,0,luminosity)
            colorName = 'blue'
            if (mode == "pulse"):
                mode = "solid"
        elif (command.lower().startswith('hey')):
            print('go away')
        elif (command.lower().startswith('ping')):
            print('pong')
        print('setting',mode,'-',color)

    # apply the color
    setColor(color)

    # if in pulse mode cycle one pulse
    if (mode == "pulse"):
        length = luminosity * 2
        for i in range(length):
            time.sleep(pulseTimeout)
            modulus = int(i % luminosity)
            if i < luminosity:
                # darkening first
                newbrightness = luminosity - modulus
            else:
                # then brightening
                newbrightness = 0 + modulus
            newColor = replace(color, luminosity, newbrightness)
            setColor(newColor)

    # lets pace ourselves here
    time.sleep(0.1)
