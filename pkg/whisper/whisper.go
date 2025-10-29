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

package whisper

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"net"
	"strconv"
	"sync"
)

// Susurrus represents a message with ID, length, and data
type Susurrus struct {
	Dest   uint32 // destination ID
	Source uint32 // source ID
	Cmd    uint16
	Nonce  uint32
	Length uint16
	Data   []byte
}

// Whisper handles multicast or unicast susurrus sending and receiving
// By default joins 224.0.6.14, but can be configured
// Received susurri are pushed to the channel
// Usage: w := NewWhisper(); w.Listen(); w.Send(...); <-w.RecvChan

type Whisper struct {
	groupAddr string
	port      int
	recvConn  *net.UDPConn // for receiving
	sendConn  *net.UDPConn // for sending
	sendAddr  *net.UDPAddr // target address for sending
	RecvChan  chan Susurrus
	closeOnce sync.Once
	closed    chan struct{}
	ID        uint32
	isUnicast bool       // flag to indicate unicast mode
	listening bool       // flag to indicate if Listen() has been called
	listenMux sync.Mutex // protects listening flag
}

// NewWhisper creates a Whisper with default group and port
func NewWhisper() (*Whisper, error) {
	w, err := NewWhisperWithOptions("224.0.6.14", 0x614)
	if err != nil {
		return nil, err
	}

	w.ID = 0

	err = w.Listen()

	if err != nil {
		return nil, err
	}

	return w, nil
}

// NewWhisperWithOptions allows custom group/address and port
// If groupAddr is a unicast address (not in multicast range), operates in unicast mode
func NewWhisperWithOptions(groupAddr string, port int) (*Whisper, error) {
	ip := net.ParseIP(groupAddr)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address: %s", groupAddr)
	}

	// Determine if this is multicast or unicast
	isMulticast := ip.IsMulticast()

	var err error

	// Create sending connection
	sendAddr := &net.UDPAddr{
		IP:   ip,
		Port: port,
	}
	sendConn, err := net.DialUDP("udp", nil, sendAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create send connection: %v", err)
	}

	fmt.Printf("Setup complete - sending to %s:%d\n", groupAddr, port)

	w := &Whisper{
		groupAddr: groupAddr,
		port:      port,
		sendConn:  sendConn,
		sendAddr:  sendAddr,
		RecvChan:  make(chan Susurrus, 16),
		closed:    make(chan struct{}),
		ID:        0,
		isUnicast: !isMulticast,
		listening: false,
	}
	return w, nil
}

// Listen starts the receive loop for incoming susurri
// This should be called once after creating a Whisper
func (w *Whisper) Listen() error {
	w.listenMux.Lock()
	defer w.listenMux.Unlock()

	if w.listening {
		return nil // Already listening
	}

	var conn *net.UDPConn
	var err error

	if w.isUnicast {
		// Unicast: listen on all interfaces
		addr := &net.UDPAddr{
			IP:   net.IPv4zero,
			Port: w.port,
		}

		fmt.Printf("Setting up unicast on %s\n", addr.String())

		conn, err = net.ListenUDP("udp", addr)
		if err != nil {
			return fmt.Errorf("failed to listen on unicast: %v", err)
		}
	} else {
		// Multicast: join the group
		mcastAddr, err := net.ResolveUDPAddr("udp", w.groupAddr+":"+strconv.Itoa(w.port))
		if err != nil {
			return fmt.Errorf("failed to resolve multicast address: %v", err)
		}

		fmt.Printf("Setting up multicast on %s\n", mcastAddr.String())

		conn, err = net.ListenMulticastUDP("udp", nil, mcastAddr)

		if err != nil {
			return fmt.Errorf("failed to listen on multicast: %v", err)
		}
	}

	w.recvConn = conn
	w.listening = true
	go w.recvLoop()

	return nil
}

// GetHashID returns a Whisper ID based on a hash of the provided data.
func (w *Whisper) GetHashID(data []byte) uint32 {
	return crc32.ChecksumIEEE(data)
}

// SetID sets the Whisper's ID
func (w *Whisper) SetID(id uint32) {
	w.ID = id
}

// SetHashID sets the Whisper's ID to a hash of the provided data.
func (w *Whisper) SetHashID(data []byte) {
	w.SetID(w.GetHashID(data))
}

// Send sends a susurrus unidirectionally (no reply tracking)
func (w *Whisper) Send(dest uint32, cmd uint16, data []byte) error {
	// generate a random nonce
	var nonceBytes [4]byte
	if _, err := rand.Read(nonceBytes[:]); err != nil {
		return err
	}
	nonce := binary.BigEndian.Uint32(nonceBytes[:])

	srs := Susurrus{
		Dest:   dest,
		Source: w.ID,
		Cmd:    cmd,
		Nonce:  nonce,
		Length: uint16(len(data)),
		Data:   data,
	}

	// Wire format: dest(4) source(4) cmd(2) nonce(4) length(2) data
	buf := make([]byte, 16+len(srs.Data))

	binary.BigEndian.PutUint32(buf[0:4], srs.Dest)
	binary.BigEndian.PutUint32(buf[4:8], srs.Source)
	binary.BigEndian.PutUint16(buf[8:10], srs.Cmd)
	binary.BigEndian.PutUint32(buf[10:14], srs.Nonce)
	binary.BigEndian.PutUint16(buf[14:16], srs.Length)
	copy(buf[16:], srs.Data)

	// Debug log
	// fmt.Printf("Sending susurrus: dest=0x%08X source=0x%08X Cmd=0x%04X Nonce=%d Len=%d\n", srs.Dest, srs.Source, srs.Cmd, srs.Nonce, len(srs.Data))

	_, err := w.sendConn.Write(buf)
	if err != nil {
		fmt.Printf("Send error: %v\n", err)
	}
	return err
}

// Note: replies are not supported - susurri are unidirectional

func (w *Whisper) recvLoop() {
	buf := make([]byte, 65535)
	w.recvConn.SetReadBuffer(len(buf))

	for {
		select {
		case <-w.closed:
			return
		default:
			n, _, err := w.recvConn.ReadFromUDP(buf)
			if err != nil || n < 16 {
				if err != nil {
					fmt.Printf("Receive error: %v\n", err)
				}
				continue
			}

			// fmt.Printf("Received %d bytes from %s\n", n, addr.String())

			// Wire format: dest(4) source(4) cmd(2) nonce(4) length(2) data
			dest := binary.BigEndian.Uint32(buf[0:4])
			source := binary.BigEndian.Uint32(buf[4:8])
			cmd := binary.BigEndian.Uint16(buf[8:10])
			nonce := binary.BigEndian.Uint32(buf[10:14])
			length := binary.BigEndian.Uint16(buf[14:16])
			if int(length) > n-16 {
				fmt.Printf("Invalid length: %d > %d\n", length, n-16)
				continue
			}
			data := make([]byte, length)
			copy(data, buf[16:16+length])

			srs := Susurrus{Dest: dest, Source: source, Cmd: cmd, Nonce: nonce, Length: length, Data: data}
			// fmt.Printf("Parsed susurrus: dest=0x%08X source=0x%08X Cmd=0x%04X Nonce=%d Len=%d\n", srs.Dest, srs.Source, srs.Cmd, srs.Nonce, len(srs.Data))

			// Filter out our own messages (by source)
			if srs.Source == w.ID {
				// fmt.Printf("Ignoring our own message (source=0x%08X)\n", w.ID)
				continue
			}

			if srs.Dest != 0 && srs.Dest != w.ID {
				fmt.Printf("Ignoring message not addressed to us (dest=0x%08X, our ID=0x%08X)\n", srs.Dest, w.ID)
				continue
			}

			// Unidirectional susurri: deliver to RecvChan
			// fmt.Printf("Sending to RecvChan\n")
			w.RecvChan <- srs
		}
	}
}
func (w *Whisper) Close() {
	w.closeOnce.Do(func() {
		close(w.closed)

		if w.recvConn != nil {
			w.recvConn.Close()
		}

		if w.sendConn != nil {
			w.sendConn.Close()
		}

		close(w.RecvChan)
	})
}
