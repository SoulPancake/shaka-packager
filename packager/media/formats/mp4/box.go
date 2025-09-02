// Copyright 2014 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package mp4

import (
	"github.com/SoulPancake/shaka-packager/packager/media/base"
)

// FourCC represents a four-character code used in MP4 files
type FourCC uint32

// Common FourCC values
const (
	FOURCC_ftyp FourCC = 0x66747970 // 'ftyp'
	FOURCC_moov FourCC = 0x6d6f6f76 // 'moov'
	FOURCC_mdat FourCC = 0x6d646174 // 'mdat'
	FOURCC_free FourCC = 0x66726565 // 'free'
	FOURCC_skip FourCC = 0x736b6970 // 'skip'
	FOURCC_mvhd FourCC = 0x6d766864 // 'mvhd'
	FOURCC_trak FourCC = 0x7472616b // 'trak'
	FOURCC_tkhd FourCC = 0x746b6864 // 'tkhd'
	FOURCC_mdia FourCC = 0x6d646961 // 'mdia'
	FOURCC_mdhd FourCC = 0x6d646864 // 'mdhd'
	FOURCC_hdlr FourCC = 0x68646c72 // 'hdlr'
	FOURCC_minf FourCC = 0x6d696e66 // 'minf'
	FOURCC_stbl FourCC = 0x7374626c // 'stbl'
	FOURCC_stsd FourCC = 0x73747364 // 'stsd'
	FOURCC_stts FourCC = 0x73747473 // 'stts'
	FOURCC_stsc FourCC = 0x73747363 // 'stsc'
	FOURCC_stsz FourCC = 0x7374737a // 'stsz'
	FOURCC_stco FourCC = 0x7374636f // 'stco'
	FOURCC_co64 FourCC = 0x636f3634 // 'co64'
)

// FourCCToString converts a FourCC to a string
func FourCCToString(fourcc FourCC) string {
	return string([]byte{
		byte(fourcc >> 24),
		byte(fourcc >> 16),
		byte(fourcc >> 8),
		byte(fourcc),
	})
}

// StringToFourCC converts a string to a FourCC
func StringToFourCC(s string) FourCC {
	if len(s) != 4 {
		return 0
	}
	return FourCC(s[0])<<24 | FourCC(s[1])<<16 | FourCC(s[2])<<8 | FourCC(s[3])
}

// BoxBuffer provides a buffer interface for reading/writing boxes
type BoxBuffer interface {
	ReadUint8() (uint8, bool)
	ReadUint16() (uint16, bool)
	ReadUint32() (uint32, bool)
	ReadUint64() (uint64, bool)
	ReadBytes(count int) ([]byte, bool)
	
	WriteUint8(v uint8)
	WriteUint16(v uint16)
	WriteUint32(v uint32)
	WriteUint64(v uint64)
	WriteBytes(data []byte)
}

// BoxReader provides reading functionality for boxes
type BoxReader struct {
	*base.BufferReader
	boxSize uint32
	boxType FourCC
}

// NewBoxReader creates a new BoxReader
func NewBoxReader(data []byte) *BoxReader {
	return &BoxReader{
		BufferReader: base.NewBufferReader(data),
	}
}

// ReadBox reads the next box from the stream
func (br *BoxReader) ReadBox() (Box, error) {
	// Read box size
	size, ok := br.BufferReader.Read4()
	if !ok {
		return nil, base.NewStatus(base.StatusCode_DataLoss, "failed to read box size")
	}
	
	// Read box type
	boxType, ok := br.BufferReader.Read4()
	if !ok {
		return nil, base.NewStatus(base.StatusCode_DataLoss, "failed to read box type")
	}
	
	br.boxSize = size
	br.boxType = FourCC(boxType)
	
	return br.createBox(FourCC(boxType))
}

// createBox creates a specific box based on the FourCC type
func (br *BoxReader) createBox(boxType FourCC) (Box, error) {
	switch boxType {
	case FOURCC_ftyp:
		return NewFileTypeBox(), nil
	case FOURCC_moov:
		return NewMovieBox(), nil
	case FOURCC_mdat:
		return NewMediaDataBox(), nil
	case FOURCC_free, FOURCC_skip:
		return NewFreeBox(), nil
	default:
		return NewUnknownBox(boxType), nil
	}
}

// Box defines the base ISO BMFF box interface
type Box interface {
	// Parse the mp4 box
	Parse(reader *BoxReader) error
	
	// Write the box to buffer
	Write(writer *base.BufferWriter) error
	
	// WriteHeader writes the box header to buffer
	WriteHeader(writer *base.BufferWriter) error
	
	// ComputeSize computes the size of this box
	ComputeSize() uint32
	
	// HeaderSize returns box header size in bytes
	HeaderSize() uint32
	
	// BoxType returns box type
	BoxType() FourCC
	
	// BoxSize returns the computed box size
	BoxSize() uint32
}

// BaseBox provides a base implementation for all boxes
type BaseBox struct {
	boxSize uint32
}

// HeaderSize returns the header size (8 bytes for basic box)
func (bb *BaseBox) HeaderSize() uint32 {
	return 8 // size (4) + type (4)
}

// BoxSize returns the computed box size
func (bb *BaseBox) BoxSize() uint32 {
	return bb.boxSize
}

// WriteHeader writes the box header to buffer
func (bb *BaseBox) WriteHeader(writer *base.BufferWriter) error {
	writer.AppendUint32(bb.ComputeSize())
	writer.AppendUint32(uint32(bb.BoxType()))
	return nil
}

// Write writes the complete box to buffer
func (bb *BaseBox) Write(writer *base.BufferWriter) error {
	bb.boxSize = bb.ComputeSize()
	if bb.boxSize == 0 {
		return nil // Don't write empty boxes
	}
	
	if err := bb.WriteHeader(writer); err != nil {
		return err
	}
	
	return bb.WriteInternal(writer)
}

// WriteInternal should be implemented by derived classes
func (bb *BaseBox) WriteInternal(writer *base.BufferWriter) error {
	return nil // Default implementation
}

// FullBox extends BaseBox with version and flags
type FullBox struct {
	BaseBox
	Version uint8
	Flags   uint32 // 24-bit flags
}

// HeaderSize returns the header size (12 bytes for full box)
func (fb *FullBox) HeaderSize() uint32 {
	return 12 // size (4) + type (4) + version (1) + flags (3)
}

// WriteHeader writes the full box header to buffer
func (fb *FullBox) WriteHeader(writer *base.BufferWriter) error {
	writer.AppendUint32(fb.ComputeSize())
	writer.AppendUint32(uint32(fb.BoxType()))
	writer.AppendUint8(fb.Version)
	writer.AppendNBytes(uint64(fb.Flags), 3) // 24-bit flags
	return nil
}

// Parse parses the version and flags from the reader
func (fb *FullBox) ParseVersionAndFlags(reader *BoxReader) error {
	version, ok := reader.Read1()
	if !ok {
		return base.NewStatus(base.StatusCode_DataLoss, "failed to read version")
	}
	fb.Version = version
	
	flags, ok := reader.ReadNBytesInto8(3)
	if !ok {
		return base.NewStatus(base.StatusCode_DataLoss, "failed to read flags")
	}
	fb.Flags = uint32(flags)
	
	return nil
}