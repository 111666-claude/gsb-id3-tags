// Package id3tags 解析 ID3v2 标签链。
package id3tags

import "unicode/utf16"

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
	index  map[string]int
}

func be32(b []byte) int {
	return int(b[0])<<24 | int(b[1])<<16 | int(b[2])<<8 | int(b[3])
}

func syncsafe(b []byte) int {
	return int(b[0]&0x7f)<<21 | int(b[1]&0x7f)<<14 | int(b[2]&0x7f)<<7 | int(b[3]&0x7f)
}

// Parse 解析标签链。
func Parse(data []byte, c *Counter) *Book {
	book := &Book{index: map[string]int{}}
	off := 0
	for off+10 <= len(data) && string(data[off:off+3]) == "ID3" {
		version := int(data[off+3])
		tag := &Tag{Version: version, Size: syncsafe(data[off+6 : off+10])}
		end := off + 10 + tag.Size
		if end > len(data) {
			end = len(data)
		}
		pos := off + 10
		for pos+10 <= end {
			c.Scanned++
			id := string(data[pos : pos+4])
			if id[0] == 0 {
				break
			}
			size := be32(data[pos+4 : pos+8])
			if version >= 4 {
				size = syncsafe(data[pos+4 : pos+8])
			}
			start := pos + 10
			if size < 0 || start+size > end {
				break
			}
			payload := data[start : start+size]
			if version >= 4 && data[pos+9]&0x01 != 0 && len(payload) >= 4 {
				payload = payload[4:]
			}
			f := Frame{ID: id, Size: size, Data: append([]byte(nil), payload...)}
			if id[0] == 'T' {
				f.Text = decodeText(f.Data)
			}
			tag.Frames = append(tag.Frames, f)
			if _, ok := book.index[id]; !ok {
				book.index[id] = len(book.Frames)
			}
			book.Frames = append(book.Frames, f)
			pos = start + size
		}
		book.Tags = append(book.Tags, tag)
		off = end
	}
	return book
}

// decodeText 按编码字节解码文本帧数据。
func decodeText(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	enc, body := data[0], data[1:]
	switch enc {
	case 0:
		runes := make([]rune, len(body))
		for i, b := range body {
			runes[i] = rune(b)
		}
		return string(runes)
	case 1:
		if len(body) >= 2 {
			if body[0] == 0xff && body[1] == 0xfe {
				return decodeUTF16(body[2:], true)
			}
			if body[0] == 0xfe && body[1] == 0xff {
				return decodeUTF16(body[2:], false)
			}
		}
		return decodeUTF16(body, false)
	case 2:
		return decodeUTF16(body, false)
	default:
		return string(body)
	}
}

func decodeUTF16(b []byte, littleEndian bool) string {
	u := make([]uint16, 0, len(b)/2)
	for i := 0; i+1 < len(b); i += 2 {
		if littleEndian {
			u = append(u, uint16(b[i])|uint16(b[i+1])<<8)
		} else {
			u = append(u, uint16(b[i])<<8|uint16(b[i+1]))
		}
	}
	return string(utf16.Decode(u))
}

// Find 按 ID 取帧。
func (b *Book) Find(id string, c *Counter) *Frame {
	c.Scanned++
	if i, ok := b.index[id]; ok {
		return &b.Frames[i]
	}
	return nil
}
