package id3tags

import (
	"fmt"
	"strings"
)

// Scenario 是一个固定场景。
type Scenario struct {
	Name string
	Run  func() map[string]any
}

func synchsafe(n int) []byte {
	return []byte{byte(n >> 21 & 0x7f), byte(n >> 14 & 0x7f), byte(n >> 7 & 0x7f), byte(n & 0x7f)}
}

type frameSpec struct {
	ID    string
	Data  []byte
	Flags uint16
}

// Build 拼一个标签。
func Build(version int, frames []frameSpec, pad int) []byte {
	var body []byte
	for _, f := range frames {
		body = append(body, []byte(f.ID)...)
		if version >= 4 {
			body = append(body, synchsafe(len(f.Data))...)
		} else {
			body = append(body, byte(len(f.Data)>>24), byte(len(f.Data)>>16), byte(len(f.Data)>>8), byte(len(f.Data)))
		}
		body = append(body, byte(f.Flags>>8), byte(f.Flags))
		body = append(body, f.Data...)
	}
	body = append(body, make([]byte, pad)...)
	out := []byte{'I', 'D', '3', byte(version), 0, 0}
	out = append(out, synchsafe(len(body))...)
	return append(out, body...)
}

// Samples 返回全部场景。
func Samples() []Scenario {
	return []Scenario{
		{Name: "synchsafe", Run: func() map[string]any {
			data := Build(3, []frameSpec{{ID: "TIT2", Data: make([]byte, 190)}}, 0)
			book := Parse(data, &Counter{})
			return map[string]any{"size": book.Tags[0].Size}
		}},
		{Name: "dli", Run: func() map[string]any {
			payload := make([]byte, 10)
			data := Build(4, []frameSpec{{ID: "TIT2", Flags: 0x0001, Data: append(synchsafe(10), payload...)}}, 0)
			book := Parse(data, &Counter{})
			f := book.Find("TIT2", &Counter{})
			if f == nil {
				return map[string]any{"datalen": 0}
			}
			return map[string]any{"datalen": len(f.Data)}
		}},
		{Name: "utf16", Run: func() map[string]any {
			data := Build(3, []frameSpec{{ID: "TIT2", Data: []byte{0x01, 0xff, 0xfe, 'A', 0x00, 'B', 0x00}}}, 0)
			book := Parse(data, &Counter{})
			f := book.Find("TIT2", &Counter{})
			if f == nil {
				return map[string]any{"textlen": 0}
			}
			return map[string]any{"textlen": len(f.Text)}
		}},
		{Name: "dupframe", Run: func() map[string]any {
			data := Build(3, []frameSpec{
				{ID: "TIT2", Data: []byte{0x00, 'A'}},
				{ID: "TIT2", Data: []byte{0x00, 'B'}},
			}, 0)
			book := Parse(data, &Counter{})
			f := book.Find("TIT2", &Counter{})
			if f == nil {
				return map[string]any{"isA": false}
			}
			return map[string]any{"isA": strings.HasSuffix(f.Text, "A")}
		}},
		{Name: "chain", Run: func() map[string]any {
			data := append(Build(3, []frameSpec{{ID: "TIT2", Data: []byte{0x00, 'A'}}}, 0),
				Build(3, []frameSpec{{ID: "TPE1", Data: []byte{0x00, 'B'}}}, 0)...)
			book := Parse(data, &Counter{})
			return map[string]any{"tags": len(book.Tags)}
		}},
	}
}

// Work 跑规模线场景。
func Work(n int) map[string]any {
	frames := make([]frameSpec, n)
	for i := range frames {
		frames[i] = frameSpec{ID: fmt.Sprintf("T%03d", i), Data: []byte{0x00, 'v'}}
	}
	book := Parse(Build(3, frames, 0), &Counter{})
	c := &Counter{}
	for i := 0; i < n; i++ {
		book.Find(fmt.Sprintf("T%03d", i), c)
	}
	return map[string]any{"frames": n, "scanned": c.Scanned}
}
