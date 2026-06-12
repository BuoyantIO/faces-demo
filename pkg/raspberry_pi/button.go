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
	"time"

	"github.com/warthog618/go-gpiocdev"
)

type ButtonEvent struct {
	Button     *Button
	PressCount int
	Timestamp  time.Duration
}

type Button struct {
	Name       string
	Debounce   time.Duration
	Line       *gpiocdev.Line
	PressCount int

	btnChan   chan ButtonEvent
	evtChan   chan gpiocdev.LineEvent
	state     int
	lastEvent gpiocdev.LineEvent
}

func NewButton(chip string, pin int, debounce time.Duration, name string, btnChan chan ButtonEvent) (*Button, error) {
	btn := &Button{
		Name:       name,
		Debounce:   debounce,
		PressCount: 0,
		state:      0,
		btnChan:    btnChan,
		evtChan:    make(chan gpiocdev.LineEvent),
		lastEvent: gpiocdev.LineEvent{
			Type: gpiocdev.LineEventRisingEdge,
		},
	}

	line, err := gpiocdev.RequestLine(chip, pin, gpiocdev.WithBothEdges, gpiocdev.WithPullDown,
		gpiocdev.WithEventHandler(func(evt gpiocdev.LineEvent) {
			btn.evtChan <- evt
		}))

	if err != nil {
		return nil, err
	}

	btn.Line = line
	go btn.watch()

	return btn, nil
}

func (btn *Button) Close() {
	btn.Line.Close()
}

func (btn *Button) watch() {
	events := make([]gpiocdev.LineEvent, 0, 2)

	// Wait for button presses
	for {
		select {
		case evt := <-btn.evtChan:
			if btn.lastEvent.Type == evt.Type {
				// Insert an event of the other type, 'cause we can't
				// get two of the same edge in a row!
				if evt.Type == gpiocdev.LineEventRisingEdge {
					events = append(events, gpiocdev.LineEvent{
						Type:      gpiocdev.LineEventFallingEdge,
						Timestamp: evt.Timestamp,
					})
				} else {
					events = append(events, gpiocdev.LineEvent{
						Type:      gpiocdev.LineEventRisingEdge,
						Timestamp: evt.Timestamp,
					})
				}
			}

			events = append(events, evt)

			for _, evt := range events {
				delta := evt.Timestamp - btn.lastEvent.Timestamp
				btn.lastEvent = evt

				fallingEdge := (evt.Type == gpiocdev.LineEventFallingEdge)

				// edge := "UP"

				// if fallingEdge {
				// 	edge = "DN"
				// }

				// fmt.Printf("%d -- %s (%s)\n", btn.state, edge, delta)

				if delta < btn.Debounce {
					continue // Ignore bounces
				}

				// It's been more than the debounce time.
				if fallingEdge {
					if btn.state == 0 {
						btn.state = 1
						btn.PressCount++
						// fmt.Printf("%s PRESS (%d)\n", btn.Name, btn.PressCount)
						btn.btnChan <- ButtonEvent{
							Button:     btn,
							PressCount: btn.PressCount,
							Timestamp:  btn.lastEvent.Timestamp,
						}
					}
				} else {
					if btn.state == 1 {
						btn.state = 0
						// fmt.Printf("%s RELEASE (%d)\n", btn.Name, btn.PressCount)
					}
				}
			}

			// Empty the events array
			events = events[:0]
		}
	}
}
