#!/usr/bin/env python3
# be sure to pip3 install pyserial
import serial

# off apple keyboard usb hub right side /dev/tty.tty.usbmodem1431301
# off apple keyboard usb hub left side /dev/tty.tty.usbmodem1431101
# off dell monitor to mini usb hub /dev/tty.tty.usbmodem1442101

ser = serial.Serial(
             '/dev/tty.usbmodem1431101',
             baudrate=115200,
             timeout=0.01)

ser.write(b'@pulse\r')
# x = ser.readlines()
# print("received: {}".format(x))
