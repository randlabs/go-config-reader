package preprocessor

import (
	"bytes"
	"errors"
	"io"
)

// -----------------------------------------------------------------------------

type TagType int

const (
	TagSRC TagType = iota + 1
	TagENV
)

// -----------------------------------------------------------------------------

type Processor struct {
	Data    []byte
	dataLen int // Do length caching ourselves
	idx     int // Current position
}

type TagInfo struct {
	p       *Processor
	Tag     TagType
	start   int
	end     int
	Content []byte
}

// -----------------------------------------------------------------------------

func New(data []byte) *Processor {
	return &Processor{
		Data:    data,
		dataLen: len(data),
		idx:     0,
	}
}

func (p *Processor) NextTag() (*TagInfo, error) {
	// Scan for the next tag
	for p.idx+6 < p.dataLen {
		// Check for tag
		if p.Data[p.idx] == '$' && p.Data[p.idx+1] == '{' && p.Data[p.idx+5] == ':' {
			switch p.Data[p.idx+2] {
			case 'S':
				// Check for SRC (source) tag
				if p.Data[p.idx+3] == 'R' && p.Data[p.idx+4] == 'C' {
					// Got a SRC tag
					return p.getTagInfo(TagSRC, 6)
				}

			case 'E':
				// Check for ENV (environment variable) tag
				if p.Data[p.idx+3] == 'N' && p.Data[p.idx+4] == 'V' {
					// Got an ENV tag
					return p.getTagInfo(TagENV, 6)
				}
			}
		}

		// No tag found, advance to next character
		p.idx += 1
	}
	return nil, io.EOF
}

func (p *Processor) getTagInfo(tag TagType, offset int) (*TagInfo, error) {
	ti := TagInfo{
		p:     p,
		Tag:   tag,
		start: p.idx,
	}

	// Skip tag start
	p.idx += offset

	// Keep track of embedded tags
	embeddedCounter := 0

	// Calculate tag's content length and look for terminator
	for p.idx < p.dataLen && (embeddedCounter != 0 || p.Data[p.idx] != '}') {

		switch p.Data[p.idx] {
		case '$': // Potential embedded tag
			if p.idx+5 < p.dataLen && p.Data[p.idx+1] == '{' && p.Data[p.idx+5] == ':' {
				embeddedCounter += 1
			}

		case '}': // End of an embedded tag
			embeddedCounter -= 1

		}

		// Advance to next character
		p.idx += 1
	}
	if p.idx >= p.dataLen {
		return nil, errors.New("error parsing tag")
	}

	// Set content. This is fast because slices in Go shares memory.
	ti.Content = p.Data[ti.start+offset : p.idx]

	// Skip end-of-tag character
	p.idx += 1

	// Set end of tag location
	ti.end = p.idx

	// Return tag info
	return &ti, nil
}

func (ti *TagInfo) Replace(newContent []byte) {
	// Update processor internals
	if newContent == nil {
		newContent = make([]byte, 0)
	}

	// Create new data buffer
	ti.p.Data = bytes.Join([][]byte{
		ti.p.Data[:ti.start],
		newContent,
		ti.p.Data[ti.end:],
	}, nil)

	// Recalculate data length
	ti.p.dataLen = len(ti.p.Data)

	//Set cursor position
	ti.p.idx = ti.start + len(newContent)
}
