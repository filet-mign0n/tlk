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
    // TODO need to circumvent square brackets better
	"[":        "(",
	"]":        ")",
}

type Convo struct {
	sync.Mutex
	Output  *string
	Input   *string
	oHeight *int
    oWidth  *int
	MyName  string
	LineCt  int
	f       *Friend
}

type msg struct {
    txt string
    ln  int
}

var WriteOutputStartFlag = true

func (c *Convo) WriteOutput(msg *msg) {
	c.Lock()
	c.LineCt++
    if !WriteOutputStartFlag {
        for msg.ln > *c.oWidth-2 {
            c.LineCt++
            msg.ln -= *c.oWidth-2
        }
        //msg.txt = fmt.Sprintf("\nmsg.ln=%d, width=%d", msg.ln, *c.oWidth)
    }
    if WriteOutputStartFlag {
        WriteOutputStartFlag = false
    }
	for c.LineCt > *c.oHeight-2 {
		c.rmFirstLine()
		c.LineCt--
	}
	var buffer bytes.Buffer
	buffer.WriteString(*c.Output)
	buffer.WriteString(msg.txt)
	newOutput := buffer.String()
	*c.Output = newOutput
	t.Render(t.Body)
	c.Unlock()
}

func (c *Convo) log(mode string, things ...interface{}) {
	txt := fmt.Sprint(things...)
    ln  := len(txt) + 3
	switch mode {
	case "d":
		if *debug {
			txt = "\n [d](fg-black,bg-yellow) " +
				"[" + txt + "](fg-yellow)"
            ln += 2
		} else {
			return
		}
	case "e":
        txt = "\n [!](fg-black,bg-red) " +
            "[" + txt + "](fg-red)"
        ln += 2
	case "s":
		txt = "\n [$ " + txt + "](fg-green)"
	default:
		txt = "\n [$ " + txt + "](fg-cyan)"
	}

    msg := &msg{
        txt: txt,
        ln : ln,
    }
	c.WriteOutput(msg)
}

func (c *Convo) chat(txt string) {
	prompt := "\n [@" + c.f.name + " ](fg-blue)"
	txt = prompt + txt
    msg := &msg{
        txt: txt,
        ln : len(c.f.name+txt)+3,
    }
	c.WriteOutput(msg)
}

func (c *Convo) inputSubmit() {
	if *c.Input == " " {
		return
	}
	if c.f != nil && c.f.conn != nil {
		c.f.out <- *c.Input
	} else {
		convo.log("d", "tui inputSubmit no c.f or c.f.conn")
	}
	prompt := "\n [@" + c.MyName + " ](fg-red)"
	txt := prompt + *c.Input
    msg := &msg {
        txt: txt,
        ln : len(c.MyName + *c.Input)+3,
    }
    c.Lock()
	*c.Input = ""
    c.Unlock()
	c.WriteOutput(msg)
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


// TODO remove not per line break, but also len
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
        oWidth:  &ob.Width,
		MyName:  "frog",
		LineCt:  13,
	}
	t.Body.Align()
	t.Render(t.Body)

	// when the window resizes, the grid must adopt to the new size.
	t.Handle("/sys/wnd/resize", func(t.Event) {
		// Update the heights of output box.
		ob.Height = t.TermHeight() - ih
        ob.Width = t.TermWidth()
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
