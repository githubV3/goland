package main

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"github.com/nsf/tulib"
	"log"
	"sync"
)

type KeyHandler func(ev termbox.Event)

type Terminal struct {
	EventChan chan termbox.Event

	AltInputHandler InputHandler
	AltChan         chan termbox.Event

	runehandlers map[rune]KeyHandler
	keyhandlers  map[termbox.Key]KeyHandler

	m sync.Mutex
}

func (t *Terminal) Start() error {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}

	t.EventChan = make(chan termbox.Event)

	// event generator
	go func(e chan termbox.Event) {
		for {
			e <- termbox.PollEvent()
		}
	}(t.EventChan)

	t.runehandlers = make(map[rune]KeyHandler)
	t.keyhandlers = make(map[termbox.Key]KeyHandler)

	return nil
}

func (t *Terminal) End() {
	termbox.Close()
}

func (t *Terminal) Clear() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
}

func (t *Terminal) Flush() {
	termbox.Flush()
}

// Change the panel which handles input.
// This should probably be push/pop.
func (t *Terminal) SetInputHandler(ih InputHandler) (old InputHandler) {
	t.m.Lock()
	defer t.m.Unlock()

	log.Printf("Changing InputHandler")

	old = t.AltInputHandler
	t.AltInputHandler = ih

	return
}

func (t *Terminal) RunInputHandlers() error {
	select {
	case ev := <-t.EventChan:

		switch ev.Type {
		case termbox.EventKey:
			log.Printf("Keypress: %s", tulib.KeyToString(ev.Key, ev.Ch, ev.Mod))

			if t.AltInputHandler != nil {
				log.Printf("Diverting keypress")
				t.AltInputHandler.HandleInput(ev)
			} else {

				if ev.Ch != 0 { // this is a character
					if handler, ok := t.runehandlers[ev.Ch]; ok {
						handler(ev)
					}
				} else {
					if handler, ok := t.keyhandlers[ev.Key]; ok {
						handler(ev)
					}
				}
			}
		case termbox.EventResize:
			// handle resize event
			t.Resize(ev.Width, ev.Height)

		case termbox.EventError:
			return fmt.Errorf("Terminal: EventError: %s", ev.Err)
		}

	default:
	}

	return nil
}

// Resize the terminal
func (t *Terminal) Resize(neww, newh int) {
}

func (t *Terminal) HandleRune(r rune, h KeyHandler) {
	t.m.Lock()
	defer t.m.Unlock()
	t.runehandlers[r] = h
}

func (t *Terminal) HandleKey(k termbox.Key, h KeyHandler) {
	t.m.Lock()
	defer t.m.Unlock()
	t.keyhandlers[k] = h
}

func (t *Terminal) PrintCell(x, y int, ch termbox.Cell) {
	termbox.SetCell(x, y, ch.Ch, ch.Fg, ch.Bg)
}

func (t *Terminal) Print(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x++
	}
}

func (t *Terminal) Printf(x, y int, fg, bg termbox.Attribute, format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	t.Print(x, y, fg, bg, s)
}
