// Copyright 2014 Google LLC. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package mp4

import (
	"github.com/SoulPancake/shaka-packager/packager/media/base"
)

// FileTypeBox represents the 'ftyp' box
type FileTypeBox struct {
	BaseBox
	MajorBrand       FourCC
	MinorVersion     uint32
	CompatibleBrands []FourCC
}

// NewFileTypeBox creates a new FileTypeBox
func NewFileTypeBox() *FileTypeBox {
	return &FileTypeBox{}
}

// BoxType returns the box type
func (ftyp *FileTypeBox) BoxType() FourCC {
	return FOURCC_ftyp
}

// Parse parses the file type box
func (ftyp *FileTypeBox) Parse(reader *BoxReader) error {
	majorBrand, ok := reader.Read4()
	if !ok {
		return base.NewStatus(base.StatusCode_DataLoss, "failed to read major brand")
	}
	ftyp.MajorBrand = FourCC(majorBrand)
	
	minorVersion, ok := reader.Read4()
	if !ok {
		return base.NewStatus(base.StatusCode_DataLoss, "failed to read minor version")
	}
	ftyp.MinorVersion = minorVersion
	
	// Read compatible brands
	remaining := reader.Remaining()
	numBrands := remaining / 4
	ftyp.CompatibleBrands = make([]FourCC, numBrands)
	
	for i := 0; i < numBrands; i++ {
		brand, ok := reader.Read4()
		if !ok {
			return base.NewStatus(base.StatusCode_DataLoss, "failed to read compatible brand")
		}
		ftyp.CompatibleBrands[i] = FourCC(brand)
	}
	
	return nil
}

// ComputeSize computes the size of this box
func (ftyp *FileTypeBox) ComputeSize() uint32 {
	size := ftyp.HeaderSize() + 8 // major brand + minor version
	size += uint32(len(ftyp.CompatibleBrands) * 4)
	ftyp.boxSize = size
	return size
}

// WriteInternal writes the file type box data
func (ftyp *FileTypeBox) WriteInternal(writer *base.BufferWriter) error {
	writer.AppendUint32(uint32(ftyp.MajorBrand))
	writer.AppendUint32(ftyp.MinorVersion)
	
	for _, brand := range ftyp.CompatibleBrands {
		writer.AppendUint32(uint32(brand))
	}
	
	return nil
}

// MovieBox represents the 'moov' box (container)
type MovieBox struct {
	BaseBox
	Children []Box
}

// NewMovieBox creates a new MovieBox
func NewMovieBox() *MovieBox {
	return &MovieBox{
		Children: make([]Box, 0),
	}
}

// BoxType returns the box type
func (moov *MovieBox) BoxType() FourCC {
	return FOURCC_moov
}

// Parse parses the movie box and its children
func (moov *MovieBox) Parse(reader *BoxReader) error {
	for reader.Remaining() > 0 {
		child, err := reader.ReadBox()
		if err != nil {
			return err
		}
		if child != nil {
			moov.Children = append(moov.Children, child)
		}
	}
	return nil
}

// ComputeSize computes the size of this box including children
func (moov *MovieBox) ComputeSize() uint32 {
	size := moov.HeaderSize()
	for _, child := range moov.Children {
		size += child.ComputeSize()
	}
	moov.boxSize = size
	return size
}

// WriteInternal writes the movie box children
func (moov *MovieBox) WriteInternal(writer *base.BufferWriter) error {
	for _, child := range moov.Children {
		if err := child.Write(writer); err != nil {
			return err
		}
	}
	return nil
}

// MediaDataBox represents the 'mdat' box
type MediaDataBox struct {
	BaseBox
	Data []byte
}

// NewMediaDataBox creates a new MediaDataBox
func NewMediaDataBox() *MediaDataBox {
	return &MediaDataBox{}
}

// BoxType returns the box type
func (mdat *MediaDataBox) BoxType() FourCC {
	return FOURCC_mdat
}

// Parse parses the media data box
func (mdat *MediaDataBox) Parse(reader *BoxReader) error {
	remaining := reader.Remaining()
	if remaining > 0 {
		data, ok := reader.ReadToVector(remaining)
		if !ok {
			return base.NewStatus(base.StatusCode_DataLoss, "failed to read media data")
		}
		mdat.Data = data
	}
	return nil
}

// ComputeSize computes the size of this box
func (mdat *MediaDataBox) ComputeSize() uint32 {
	size := mdat.HeaderSize() + uint32(len(mdat.Data))
	mdat.boxSize = size
	return size
}

// WriteInternal writes the media data
func (mdat *MediaDataBox) WriteInternal(writer *base.BufferWriter) error {
	writer.AppendBytes(mdat.Data)
	return nil
}

// FreeBox represents 'free' or 'skip' boxes
type FreeBox struct {
	BaseBox
	boxType FourCC
	Data    []byte
}

// NewFreeBox creates a new FreeBox
func NewFreeBox() *FreeBox {
	return &FreeBox{
		boxType: FOURCC_free,
	}
}

// NewSkipBox creates a new SkipBox
func NewSkipBox() *FreeBox {
	return &FreeBox{
		boxType: FOURCC_skip,
	}
}

// BoxType returns the box type
func (free *FreeBox) BoxType() FourCC {
	return free.boxType
}

// Parse parses the free box
func (free *FreeBox) Parse(reader *BoxReader) error {
	remaining := reader.Remaining()
	if remaining > 0 {
		data, ok := reader.ReadToVector(remaining)
		if !ok {
			return base.NewStatus(base.StatusCode_DataLoss, "failed to read free box data")
		}
		free.Data = data
	}
	return nil
}

// ComputeSize computes the size of this box
func (free *FreeBox) ComputeSize() uint32 {
	size := free.HeaderSize() + uint32(len(free.Data))
	free.boxSize = size
	return size
}

// WriteInternal writes the free box data
func (free *FreeBox) WriteInternal(writer *base.BufferWriter) error {
	writer.AppendBytes(free.Data)
	return nil
}

// UnknownBox represents an unknown box type
type UnknownBox struct {
	BaseBox
	boxType FourCC
	Data    []byte
}

// NewUnknownBox creates a new UnknownBox
func NewUnknownBox(boxType FourCC) *UnknownBox {
	return &UnknownBox{
		boxType: boxType,
	}
}

// BoxType returns the box type
func (unknown *UnknownBox) BoxType() FourCC {
	return unknown.boxType
}

// Parse parses the unknown box by reading all remaining data
func (unknown *UnknownBox) Parse(reader *BoxReader) error {
	remaining := reader.Remaining()
	if remaining > 0 {
		data, ok := reader.ReadToVector(remaining)
		if !ok {
			return base.NewStatus(base.StatusCode_DataLoss, "failed to read unknown box data")
		}
		unknown.Data = data
	}
	return nil
}

// ComputeSize computes the size of this box
func (unknown *UnknownBox) ComputeSize() uint32 {
	size := unknown.HeaderSize() + uint32(len(unknown.Data))
	unknown.boxSize = size
	return size
}

// WriteInternal writes the unknown box data
func (unknown *UnknownBox) WriteInternal(writer *base.BufferWriter) error {
	writer.AppendBytes(unknown.Data)
	return nil
}