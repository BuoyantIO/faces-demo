// SPDX-FileCopyrightText: 2025 Buoyant Inc.
// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2022-2025 Buoyant Inc.
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

package raspberry_pi

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/BuoyantIO/faces-demo/v2/pkg/faces"
	"github.com/BuoyantIO/faces-demo/v2/pkg/utils"
	"github.com/warthog618/go-gpiocdev"
)

type HardwareStuff struct {
	logger      *slog.Logger
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

func NewAutomaticHardwareStuff() (*HardwareStuff, error) {
	rotaryAPin := utils.IntFromEnv("ROTARY_A_PIN", 5)
	rotaryBPin := utils.IntFromEnv("ROTARY_B_PIN", 6)
	buttonPin := utils.IntFromEnv("BUTTON_PIN", 4)
	ledGreenPin := utils.IntFromEnv("LED_GREEN_PIN", 19)
	ledRedPin := utils.IntFromEnv("LED_RED_PIN", 13)

	return NewHardwareStuff(rotaryAPin, rotaryBPin, buttonPin, ledGreenPin, ledRedPin)
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

		logger: slog.Default().With("provider", "hw"),

		button: btn,
		rotary: rotary,
		leds: map[string]*gpiocdev.Line{
			"red":   redLED,
			"green": greenLED,
		},

		btnChan:    btnChan,
		rotaryChan: rotaryChan,
	}

	hw.Infof("rotaryAPin %d", hw.rotaryAPin)
	hw.Infof("rotaryBPin %d", hw.rotaryBPin)
	hw.Infof("buttonPin %d", hw.buttonPin)
	hw.Infof("ledGreenPin %d", hw.ledGreenPin)
	hw.Infof("ledRedPin %d", hw.ledRedPin)

	return hw, nil
}

func (hw *HardwareStuff) Close() {
	hw.button.Close()
	hw.rotary.Close()
	hw.leds["red"].Close()
	hw.leds["green"].Close()
}

func (hw *HardwareStuff) Infof(format string, args ...interface{}) {
	hw.logger.Info("hw: " + fmt.Sprintf(format, args...))
}

func (hw *HardwareStuff) Debugf(format string, args ...interface{}) {
	hw.logger.Debug("hw: " + fmt.Sprintf(format, args...))
}

func (hw *HardwareStuff) Warnf(format string, args ...interface{}) {
	hw.logger.Warn("hw: " + fmt.Sprintf(format, args...))
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

			hw.Infof("error fraction now %d\n", hw.serverErrorFraction)
		}
	}()

	go func() {
		for {
			<-hw.btnChan

			hw.serverLatched = !hw.serverLatched

			hw.Infof("hardware: latched now %v\n", hw.serverLatched)
		}
	}()
}

func (hw *HardwareStuff) Updater(bprv *faces.BaseProvider) {
	bprv.Lock()
	defer bprv.Unlock()

	if bprv.ErrorFraction() != hw.serverErrorFraction {
		bprv.SetErrorFraction(hw.serverErrorFraction)
		bprv.Infof("Updater: set errorFraction to %d", bprv.ErrorFraction())
	}

	if bprv.IsLatched() != hw.serverLatched {
		bprv.SetLatched(hw.serverLatched)
		bprv.Infof("Updater: set latched to %v", bprv.IsLatched())
	}
}

func (hw *HardwareStuff) PreHook(bprv *faces.BaseProvider, prvReq *faces.ProviderRequest, rstat *faces.BaseRequestStatus) bool {
	if rstat.IsErrored() || rstat.IsRateLimited() {
		hw.ledOn("red")
	} else {
		hw.ledOn("green")
	}

	return true
}

func (hw *HardwareStuff) PostHook(bprv *faces.BaseProvider, prvReq *faces.ProviderRequest, rstat *faces.BaseRequestStatus) bool {
	hw.ledOff("red")
	hw.ledOff("green")

	return true
}
