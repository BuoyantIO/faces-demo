// SPDX-FileCopyrightText: 2024 Buoyant Inc.
// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2022-2024 Buoyant Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.  You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"time"

	"github.com/warthog618/go-gpiocdev"
)

type HardwareStuff struct {
	rotaryAPin  int
	rotaryBPin  int
	buttonPin   int
	ledGreenPin int
	ledRedPin   int

	button *Button
	rotary *RotaryEncoder

	leds map[string]*gpiocdev.Line

	btnChan    chan ButtonEvent
	rotaryChan chan RotaryEvent

	serverErrorFraction int
	serverLatched       bool
}

func NewHardwareStuff(rotaryAPin, rotaryBPin, buttonPin, ledGreenPin, ledRedPin int) (*HardwareStuff, error) {
	btnChan := make(chan ButtonEvent)
	btn, err := NewButton("gpiochip0", buttonPin, 50*time.Millisecond, "Button", btnChan)

	if err != nil {
		return nil, fmt.Errorf("could not create button on line %d: %s", buttonPin, err)
	}

	rotaryChan := make(chan RotaryEvent)
	rotary, err := NewRotaryEncoder("gpiochip0", rotaryAPin, rotaryBPin, 10*time.Millisecond, "Encoder1", rotaryChan)

	if err != nil {
		btn.Close()
		return nil, fmt.Errorf("could not create rotary encoder on lines %d and %d: %s", rotaryAPin, rotaryBPin, err)
	}

	redLED, err := gpiocdev.RequestLine("gpiochip0", ledRedPin, gpiocdev.AsOutput(1), gpiocdev.LineDrivePushPull)

	if err != nil {
		btn.Close()
		rotary.Close()
		redLED.Close()
		return nil, fmt.Errorf("could not create red LED on line %d: %s", ledRedPin, err)
	}

	greenLED, err := gpiocdev.RequestLine("gpiochip0", ledGreenPin, gpiocdev.AsOutput(1), gpiocdev.LineDrivePushPull)

	if err != nil {
		btn.Close()
		rotary.Close()
		redLED.Close()
		greenLED.Close()
		return nil, fmt.Errorf("could not create green LED on line %d: %s", ledGreenPin, err)
	}

	hw := &HardwareStuff{
		rotaryAPin:  rotaryAPin,
		rotaryBPin:  rotaryBPin,
		buttonPin:   buttonPin,
		ledGreenPin: ledGreenPin,
		ledRedPin:   ledRedPin,

		button: btn,
		rotary: rotary,
		leds: map[string]*gpiocdev.Line{
			"red":   redLED,
			"green": greenLED,
		},

		btnChan:    btnChan,
		rotaryChan: rotaryChan,
	}

	return hw, nil
}

func (hw *HardwareStuff) Close() {
	hw.button.Close()
	hw.rotary.Close()
	hw.leds["red"].Close()
	hw.leds["green"].Close()
}

// ledOn turns on the LED of the specified color.
func (hw *HardwareStuff) ledOn(color string) {
	hw.leds[color].SetValue(0)
}

// ledOff turns on the LED of the specified color.
func (hw *HardwareStuff) ledOff(color string) {
	hw.leds[color].SetValue(1)
}

func (hw *HardwareStuff) Watch(startingErrorFraction int, startingLatched bool) {
	hw.serverErrorFraction = startingErrorFraction
	hw.serverLatched = startingLatched

	go func() {
		for {
			evt := <-hw.rotaryChan

			efrac := hw.serverErrorFraction + (evt.Direction * 5)

			if efrac < 0 {
				efrac = 0
			}

			if efrac > 100 {
				efrac = 100
			}

			hw.serverErrorFraction = efrac

			fmt.Printf("hardware: error fraction now %d\n", hw.serverErrorFraction)
		}
	}()

	go func() {
		for {
			<-hw.btnChan

			hw.serverLatched = !hw.serverLatched

			fmt.Printf("hardware: latched now %v\n", hw.serverLatched)
		}
	}()
}
