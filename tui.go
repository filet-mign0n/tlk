package main

import (
	"bytes"
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"

	t "github.com/gizak/termui"
)

const (
	lw       = 20
	ih       = 3
	startMsg = `  [press ctrl-c to quit](fg-red)

  @@@@@@@  @@@       @@@  @@@
  @@@@@@@  @@@       @@@  @@@
    @@!    @@!       @@!  !@@
    !@!    !@!       !@!  @!!
    @!!    @!!       @!@@!@!
    !!!    !!!       !!@!!!
    !!:    !!:       !!: :!!
    :!:     :!:      :!:  !:!
     ::     :: ::::   ::  :::
     :     : :: : :   :   :::
`
)

var convo *Convo

var specialKeys = map[string]string{
	"C-8":      "_del_",
	"<tab>":    "    ",
	"<space>":  " ",
	"<escape>": "",
	"<left>":   "",
	"<up>":     "",
	"<right>":  "",
	"<down>":   "",
	"[":        "",
	"]":        "",
}

type Convo struct {
	sync.Mutex
	Output  *string
	Input   *string
	oHeight *int
	MyName  string
	LineCt  int
	f       *Friend
}

func (c *Convo) WriteOutput(msg string) {
	c.Lock()
	c.LineCt++
	if c.LineCt > *c.oHeight-2 {
		c.rmFirstLine()
		c.LineCt--
	}
	var buffer bytes.Buffer
	buffer.WriteString(*c.Output)
	buffer.WriteString(msg)
	newOutput := buffer.String()
	*c.Output = newOutput
	c.Unlock()

	t.Render(t.Body)
}

func (c *Convo) log(mode string, things ...interface{}) {
	msg := fmt.Sprint(things...)
	switch mode {
	case "d":
		if *debug {
			msg = "\n [d](fg-black,bg-yellow) " +
				"[" + msg + "](fg-yellow)"
		} else {
			return
		}
	case "e":
		msg = "\n [$ " + msg + "](fg-red)"
	case "s":
		msg = "\n [$ " + msg + "](fg-green)"
	default:
		msg = "\n [$ " + msg + "](fg-cyan)"
	}
	c.WriteOutput(msg)
}

func (c *Convo) chat(msg string) {
	prompt := "\n [@" + c.f.name + " ](fg-blue)"
	msg = prompt + msg
	c.WriteOutput(msg)
}

func (c *Convo) inputSubmit() {
	if *c.Input == " " {
		return
	}
	if c.f != nil && c.f.conn != nil {
		c.f.out <- *c.Input
	} else {
		convo.log("d", "tui inputSubmit no c.f or c.f.conn?")
	}
	prompt := "\n [@" + c.MyName + " ](fg-red)"
	newChat := prompt + *c.Input
	*c.Input = ""
	c.WriteOutput(newChat)
}

func (c *Convo) keyInput(key string) {
	var buffer bytes.Buffer
	key = c.handleKey(key)
	if key == "_del_" {
		if last := len(*c.Input) - 1; last >= 0 {
			*c.Input = (*c.Input)[:last]
			t.Render(t.Body)
			return
		}
		return
	}
	buffer.WriteString(*c.Input)
	buffer.WriteString(key)
	newInput := buffer.String()
	*c.Input = newInput

	t.Render(t.Body)
}

func (c *Convo) handleKey(key string) string {
	if sKey, ok := specialKeys[key]; ok {
		return sKey
	}
	return key
}

func (c *Convo) rmFirstLine() {
	if idx := strings.Index(*c.Output, "\n"); idx != -1 {
		output := *c.Output
		*c.Output = output[idx+1:]
		t.Render(t.Body)
	}
}

func tui(wg *sync.WaitGroup) {
	defer wg.Done()
	err := t.Init()
	if err != nil {
		log.Fatalln("cannot initialize termui")
	}
	defer t.Close()
	th := t.TermHeight()

	// input block
	ib := t.NewPar("")
	ib.Height = ih
	ib.BorderLabelFg = t.ColorYellow
	ib.BorderFg = t.ColorYellow
	ib.TextFgColor = t.ColorWhite

	// output block.
	ob := t.NewPar(startMsg)
	ob.Height = th - ih
	ob.BorderLabel = "tlk"
	ob.BorderLabelFg = t.ColorCyan
	ob.BorderFg = t.ColorCyan
	ob.TextFgColor = t.ColorWhite

	t.Body.AddRows(
		t.NewRow(
			t.NewCol(12, 0, ob, ib)))

	convo = &Convo{
		Output:  &ob.Text,
		Input:   &ib.Text,
		oHeight: &ob.Height,
		MyName:  "frog",
		LineCt:  13,
	}
	t.Body.Align()
	t.Render(t.Body)

	// when the window resizes, the grid must adopt to the new size.
	t.Handle("/sys/wnd/resize", func(t.Event) {
		// Update the heights of output box.
		ob.Height = t.TermHeight() - ih
		t.Body.Width = t.TermWidth()
		t.Body.Align()
		t.Render(t.Body)
	})
	// way to test the off channel w/o killing tui session
	t.Handle("/sys/kbd/C-d", func(t.Event) {
		if *debug {
			close(off)
		}
	})
	// Ctrl-C stops the TUI event loop and kills all goroutines
	t.Handle("/sys/kbd/C-c", func(t.Event) {
		close(off)
		t.StopLoop()
		return
	})
	t.Handle("/sys/kbd/<enter>", func(t.Event) {
		if len(*convo.Input) > 0 {
			convo.inputSubmit()
		}
	})
	t.Handle("/sys/kbd", func(e t.Event) {
		// handle all other key presses
		v := reflect.ValueOf(e.Data)
		s := v.Field(0)
		convo.keyInput(s.String())
	})
	t.Loop()
}
