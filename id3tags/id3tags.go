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

// syncsafe 解码同步安全整数：每字节只有低 7 位有效。
func syncsafe(b []byte) int {
	return int(b[0]&0x7f)<<21 | int(b[1]&0x7f)<<14 | int(b[2]&0x7f)<<7 | int(b[3]&0x7f)
}

// Parse 解析标签链：文件里可以接着放多个 ID3v2 标签，逐个解析。
func Parse(data []byte, c *Counter) *Book {
	book := &Book{index: map[string]int{}}
	off := 0
	for len(data)-off >= 10 && string(data[off:off+3]) == "ID3" {
		tag, next := parseTag(data[off:], book, c)
		if tag == nil {
			break
		}
		book.Tags = append(book.Tags, tag)
		off += next
	}
	return book
}

// parseTag 解析单个标签，返回标签与其总长度（头部加数据）。
func parseTag(data []byte, book *Book, c *Counter) (*Tag, int) {
	if len(data) < 10 || string(data[:3]) != "ID3" {
		return nil, 0
	}
	version := int(data[3])
	size := syncsafe(data[6:10])
	tag := &Tag{Version: version, Size: size}
	end := 10 + size
	if end > len(data) {
		end = len(data)
	}
	off := 10
	for off+10 <= end {
		c.Scanned++
		id := string(data[off : off+4])
		if id[0] == 0 {
			break
		}
		// 帧长度：v2.3 普通大端，v2.4 同步安全。
		var fsize int
		if version >= 4 {
			fsize = syncsafe(data[off+4 : off+8])
		} else {
			fsize = be32(data[off+4 : off+8])
		}
		flags := data[off+8 : off+10]
		start := off + 10
		if fsize < 0 || start+fsize > end {
			break
		}
		payload := data[start : start+fsize]
		// v2.4 帧标志最低位置位时，数据开头有 4 字节同步安全的数据长度指示符。
		if version >= 4 && flags[1]&0x01 != 0 {
			if len(payload) < 4 {
				break
			}
			payload = payload[4:]
		}
		f := Frame{ID: id, Size: fsize, Data: append([]byte(nil), payload...)}
		if id[0] == 'T' {
			f.Text = decodeText(f.Data)
		} else {
			f.Text = string(f.Data)
		}
		tag.Frames = append(tag.Frames, f)
		// 去重：同一个帧 ID 只保留第一次出现的那一帧。
		if _, ok := book.index[id]; !ok {
			book.index[id] = len(book.Frames)
			book.Frames = append(book.Frames, f)
		}
		off = start + fsize
	}
	return tag, 10 + size
}

// decodeText 按首字节编码解码文本帧数据，返回值不含编码字节。
func decodeText(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	body := data[1:]
	switch data[0] {
	case 0: // 拉丁
		r := make([]rune, len(body))
		for i, b := range body {
			r[i] = rune(b)
		}
		return string(r)
	case 1: // 带 BOM 的 UTF-16
		little := false
		if len(body) >= 2 {
			switch {
			case body[0] == 0xff && body[1] == 0xfe:
				little = true
				body = body[2:]
			case body[0] == 0xfe && body[1] == 0xff:
				body = body[2:]
			}
		}
		return utf16Text(body, little)
	case 2: // 大端 UTF-16
		return utf16Text(body, false)
	case 3: // UTF-8
		return string(body)
	default:
		return string(body)
	}
}

func utf16Text(b []byte, little bool) string {
	u := make([]uint16, 0, len(b)/2)
	for i := 0; i+1 < len(b); i += 2 {
		if little {
			u = append(u, uint16(b[i])|uint16(b[i+1])<<8)
		} else {
			u = append(u, uint16(b[i])<<8|uint16(b[i+1]))
		}
	}
	return string(utf16.Decode(u))
}

// Find 按 ID 取第一帧：走索引，不逐帧比较。
func (b *Book) Find(id string, c *Counter) *Frame {
	c.Scanned++
	if i, ok := b.index[id]; ok {
		return &b.Frames[i]
	}
	return nil
}
