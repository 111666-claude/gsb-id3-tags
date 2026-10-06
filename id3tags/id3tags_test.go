package id3tags

import "testing"

func TestVersionStored(t *testing.T) {
	book := Parse(Build(3, []frameSpec{{ID: "TIT2", Data: []byte{0x00, 'a'}}}, 0), &Counter{})
	if book.Tags[0].Version != 3 {
		t.Fatalf("版本应是 3：%d", book.Tags[0].Version)
	}
}

func TestFrameCount(t *testing.T) {
	book := Parse(Build(3, []frameSpec{{ID: "TIT2", Data: []byte{0x00, 'a'}}}, 0), &Counter{})
	if len(book.Tags[0].Frames) != 1 {
		t.Fatalf("应有一个帧：%d", len(book.Tags[0].Frames))
	}
}

func TestFrameID(t *testing.T) {
	book := Parse(Build(3, []frameSpec{{ID: "TIT2", Data: []byte{0x00, 'a'}}}, 0), &Counter{})
	if book.Tags[0].Frames[0].ID != "TIT2" {
		t.Fatalf("帧 ID 应是 TIT2：%q", book.Tags[0].Frames[0].ID)
	}
}

func TestFrameDataLen(t *testing.T) {
	book := Parse(Build(3, []frameSpec{{ID: "TIT2", Data: []byte("abcd")}}, 0), &Counter{})
	if len(book.Tags[0].Frames[0].Data) != 4 {
		t.Fatalf("数据长度应是 4：%d", len(book.Tags[0].Frames[0].Data))
	}
}

func TestFindMissing(t *testing.T) {
	book := Parse(Build(3, []frameSpec{{ID: "TIT2", Data: []byte{0x00, 'a'}}}, 0), &Counter{})
	if book.Find("NOPE", &Counter{}) != nil {
		t.Fatal("缺失的帧应返回 nil")
	}
}

func TestCounterCounts(t *testing.T) {
	c := &Counter{}
	Parse(Build(3, []frameSpec{{ID: "TIT2", Data: []byte{0x00, 'a'}}}, 0), c)
	if c.Scanned == 0 {
		t.Fatal("应统计比较次数")
	}
}
