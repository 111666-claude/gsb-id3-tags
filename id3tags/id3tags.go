// Package id3tags 解析 ID3v2 标签链。
package id3tags

// Counter 记录比较次数。
type Counter struct {
	Scanned int
}

// Frame 是一个帧。
type Frame struct {
	ID   string
	Size int
	Data []byte
	Text string
}

// Tag 是一个标签。
type Tag struct {
	Version int
	Size    int
	Frames  []Frame
}

// Book 是解析结果。
type Book struct {
	Tags   []*Tag
	Frames []Frame
}

func be32(b []byte) int {
	return int(b[0])<<24 | int(b[1])<<16 | int(b[2])<<8 | int(b[3])
}

// Parse 解析标签链。
func Parse(data []byte, c *Counter) *Book {
	book := &Book{}
	if len(data) < 10 || string(data[:3]) != "ID3" {
		return book
	}
	tag := &Tag{Version: int(data[3]), Size: be32(data[6:10])}
	off := 10
	for off+10 <= len(data) {
		c.Scanned++
		id := string(data[off : off+4])
		if id[0] == 0 {
			break
		}
		size := be32(data[off+4 : off+8])
		start := off + 10
		if size < 0 || start+size > len(data) {
			break
		}
		f := Frame{ID: id, Size: size, Data: append([]byte(nil), data[start:start+size]...)}
		f.Text = string(f.Data)
		tag.Frames = append(tag.Frames, f)
		book.Frames = append(book.Frames, f)
		off = start + size
	}
	book.Tags = append(book.Tags, tag)
	return book
}

// Find 按 ID 取帧。
func (b *Book) Find(id string, c *Counter) *Frame {
	var found *Frame
	for i := range b.Frames {
		c.Scanned++
		if b.Frames[i].ID == id {
			found = &b.Frames[i]
		}
	}
	return found
}
