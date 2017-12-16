package main

import (
	"bytes"
	"log"
	"net"
	"reflect"
	"strings"
	"sync"

	t "github.com/gizak/termui"
)

const (
	lw       = 20
	ih       = 3
	startMsg = "\n  [press Ctrl-C to quit](fg-red)\n"
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
	Output  *string
	Input   *string
	oHeight *int
	MyName  string
	LineCt  int
	f       *Friend
}

func (c *Convo) WriteOutput(msg string) {
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

	t.Render(t.Body)

}

func (c *Convo) log(msg string) {
	msg = "\n [$ " + msg + "](fg-green)"
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
	if c.f != nil {
		c.f.out <- *c.Input
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

func runTermui(ch chan<- net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()

	err := t.Init()
	if err != nil {
		log.Fatalln("Cannot initialize termui")
	}
	defer t.Close()

	th := t.TermHeight()

	// The input block. termui has no edit box yet, but at the time of
	// this writing, there is an open [pull request](https://github.com/gizak/termui/pull/129) for adding
	// a text input widget.
	ib := t.NewPar("")
	ib.Height = ih
	//ib.BorderLabel = "Message"
	ib.BorderLabelFg = t.ColorYellow
	ib.BorderFg = t.ColorYellow
	ib.TextFgColor = t.ColorWhite

	// The Output block.
	ob := t.NewPar(startMsg)
	ob.Height = th - ih
	ob.BorderLabel = "tlk"
	ob.BorderLabelFg = t.ColorCyan
	ob.BorderFg = t.ColorCyan
	ob.TextFgColor = t.ColorWhite

	// Now we need to create the layout. The blocks have gotten a size
	// but no position. A grid layout puts everything into place.
	// t.Body is a pre-defined grid. We add one row that contains
	// two columns.
	//
	// The grid uses a 12-column system, so we have to give a "span"
	// parameter to each column that specifies how many grid column
	// each column occupies.
	t.Body.AddRows(
		t.NewRow(
			t.NewCol(12, 0, ob, ib)))

	convo = &Convo{
		Output:  &ob.Text,
		Input:   &ib.Text,
		oHeight: &ob.Height,
		MyName:  "frog",
		LineCt:  3,
	}
	// Render the grid.
	t.Body.Align()
	t.Render(t.Body)

	// When the window resizes, the grid must adopt to the new size.
	// We use a hander func for this.
	t.Handle("/sys/wnd/resize", func(t.Event) {
		// Update the heights of output box.
		ob.Height = t.TermHeight() - ih
		t.Body.Width = t.TermWidth()
		t.Body.Align()
		t.Render(t.Body)
	})
	// We need a way out. Ctrl-C shall stop the event loop.
	t.Handle("/sys/kbd/C-c", func(t.Event) {
		if convo.f != nil {
			close(convo.f.off)
		}
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
