// Command id3tags 是 ID3 标签的场景入口。
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"example.com/id3-tags/id3tags"
)

func emit(v any) {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}

func main() {
	sample := flag.String("sample", "", "synchsafe|dli|utf16|dupframe|chain|work")
	flag.Parse()

	for _, s := range id3tags.Samples() {
		if s.Name == *sample {
			emit(s.Run())
			return
		}
	}
	if *sample == "work" {
		emit(id3tags.Work(800))
		return
	}
	fmt.Fprintln(os.Stderr, "未知场景："+*sample)
	os.Exit(2)
}
