// SPDX-FileCopyrightText: 2024 Buoyant Inc.
// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2022-2024 Buoyant Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.  You may obtain
// a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/warthog618/go-gpiocdev"
)

type RotaryEvent struct {
	Rotary    *RotaryEncoder
	Position  int
	Direction int
	Timestamp time.Duration
}

type rotaryDebugEvent struct {
	timestamp time.Duration
	what      string
	moving    bool
	direction int
	state     int
}

type RotaryEncoder struct {
	Name     string
	Debounce time.Duration
	LineA    *gpiocdev.Line
	LineB    *gpiocdev.Line
	Position int

	Debug bool

	lock sync.Mutex

	evtChan       chan RotaryEvent
	gpioChan      chan gpiocdev.LineEvent
	debounceChan  chan interface{}
	debounceTimer *time.Timer
	state         int
	moving        bool
	direction     int

	debugEvents []rotaryDebugEvent
}

func NewRotaryEncoder(chip string, pinA int, pinB int, debounce time.Duration, name string, evtChan chan RotaryEvent) (*RotaryEncoder, error) {
	enc := &RotaryEncoder{
		Name:         name,
		Debounce:     debounce,
		Position:     0,
		state:        3, // Assume we're starting with the knob stationary with both pins high.
		evtChan:      evtChan,
		debounceChan: make(chan interface{}),
		gpioChan:     make(chan gpiocdev.LineEvent),

		debugEvents: make([]rotaryDebugEvent, 0, 10),
	}

	lineA, err := gpiocdev.RequestLine(chip, pinA, gpiocdev.WithBothEdges, gpiocdev.WithPullUp,
		gpiocdev.WithEventHandler(func(evt gpiocdev.LineEvent) {
			enc.gpioChan <- evt
		}))

	if err != nil {
		return nil, err
	}

	lineB, err := gpiocdev.RequestLine(chip, pinB, gpiocdev.WithBothEdges, gpiocdev.WithPullUp,
		gpiocdev.WithEventHandler(func(evt gpiocdev.LineEvent) {
			enc.gpioChan <- evt
		}))

	if err != nil {
		return nil, err
	}

	enc.LineA = lineA
	enc.LineB = lineB
	enc.state = enc.ReadState()
	go enc.watch()

	return enc, nil
}

func (enc *RotaryEncoder) ReadState() int {
	stateA, _ := enc.LineA.Value()
	stateB, _ := enc.LineB.Value()

	return (stateA << 1) | stateB
}

func (enc *RotaryEncoder) Close() {
	enc.LineA.Close()
	enc.LineB.Close()
}

func (enc *RotaryEncoder) postDebugEvent(what string) {
	if !enc.Debug {
		return
	}

	enc.debugEvents = append(enc.debugEvents, rotaryDebugEvent{
		timestamp: time.Duration(time.Now().UnixNano()),
		what:      what,
		moving:    enc.moving,
		direction: enc.direction,
		state:     enc.state,
	})

	// enc.printDebugEvent(enc.debugEvents[len(enc.debugEvents)-1], 0)
}

func (enc *RotaryEncoder) printDebugEvent(evt rotaryDebugEvent, lastTime time.Duration) {
	delta := evt.timestamp - lastTime
	lastTime = evt.timestamp

	mstr := "-"

	if evt.moving {
		mstr = "M"
	}

	dstr := "-"

	if evt.direction < 0 {
		dstr = "<"
	} else if evt.direction > 0 {
		dstr = ">"
	}

	fmt.Printf("%8d: %02b %s%s %s\n", delta, evt.state, mstr, dstr, evt.what)
}

func (enc *RotaryEncoder) DumpDebugEvents() {
	lastTime := time.Duration(0)

	for _, evt := range enc.debugEvents {
		enc.printDebugEvent(evt, lastTime)
		lastTime = evt.timestamp
	}

	enc.debugEvents = make([]rotaryDebugEvent, 0, 10)
}

func (enc *RotaryEncoder) clearDebounceTimer() {
	enc.lock.Lock()
	defer enc.lock.Unlock()

	if enc.debounceTimer != nil {
		enc.debounceTimer.Stop()
		enc.debounceTimer = nil
		enc.postDebugEvent("stop debounce timer")
	} else {
		enc.postDebugEvent("no debounce timer")
	}
}

func (enc *RotaryEncoder) startDebounceTimer() {
	enc.lock.Lock()
	defer enc.lock.Unlock()

	timer := time.NewTimer(enc.Debounce)
	go func() {
		<-timer.C
		enc.debounceChan <- struct{}{}
	}()

	enc.debounceTimer = timer
}

func (enc *RotaryEncoder) watch() {
	// Wait for transitions presses
	for {
		select {
		case <-enc.gpioChan:
			curState := enc.ReadState()

			enc.postDebugEvent(fmt.Sprintf("GPIO event, new state %02b", curState))

			enc.clearDebounceTimer()

			if enc.moving {
				// We're moving. If we're in state 3, start the debounce timer so that we
				// can tell when we're done moving.
				if curState == 0b11 {
					enc.startDebounceTimer()
					enc.postDebugEvent("hit state 3, starting timer")
				}

				enc.state = curState
			} else {
				// Not moving, so let's figure out what's up.
				if curState != enc.state {
					// We're not already moving, so we're starting to move,
					// and this first transition will tell us which direction
					// we're moving.
					enc.moving = true

					// We're treating it as kind of axiomatic that they can't flip from
					// 11 to 00 without us seeing a 01 or 10 first.
					if curState == 0b01 {
						enc.direction = -1
					} else {
						enc.direction = 1
					}

					enc.postDebugEvent(fmt.Sprintf("starting move %d", enc.direction))
					// fmt.Printf("%s starting move %d\n", enc.Name, enc.direction)

					enc.state = curState
				}
			}

		case <-enc.debounceChan:
			// We've been in state 3 for the debounce time, so we're done moving.
			if !enc.moving {
				// Whut.
				enc.postDebugEvent("debounce fired but not moving?")
				// fmt.Printf("%s debounce timer expired, but we're not moving\n", enc.Name)
			} else {
				enc.postDebugEvent("debounce fired, done moving")
				enc.Position += enc.direction

				enc.evtChan <- RotaryEvent{
					Rotary:    enc,
					Position:  enc.Position,
					Direction: enc.direction,
					Timestamp: time.Duration(time.Now().UnixNano()),
				}

				enc.moving = false
				enc.direction = 0
				// fmt.Printf("%s done moving %d\n", enc.Name, enc.direction)
			}
		}
	}
}
