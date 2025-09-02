// Package mp4 provides MP4/ISO-BMFF container format support
package mp4

import (
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

// Box represents an MP4 box (atom)
type Box interface {
	Type() [4]byte
	Size() uint64
	WriteTo(w io.Writer) error
	ReadFrom(r io.Reader) error
}

// BaseBox provides common functionality for all boxes
type BaseBox struct {
	BoxSize uint32
	BoxType [4]byte
}

// Type returns the box type
func (b *BaseBox) Type() [4]byte {
	return b.BoxType
}

// Size returns the box size
func (b *BaseBox) Size() uint64 {
	return uint64(b.BoxSize)
}

// WriteTo implements the Box interface for BaseBox
func (b *BaseBox) WriteTo(w io.Writer) error {
	// Base implementation - write size and type only
	return binary.Write(w, binary.BigEndian, b)
}

// GenericBox represents a generic MP4 box
type GenericBox struct {
	BaseBox
	Data []byte
}

// ReadFrom implements the Box interface for GenericBox
func (g *GenericBox) ReadFrom(r io.Reader) error {
	// Read the data
	g.Data = make([]byte, g.BoxSize-8)
	_, err := io.ReadFull(r, g.Data)
	return err
}

// WriteTo implements the Box interface for GenericBox
func (g *GenericBox) WriteTo(w io.Writer) error {
	if err := binary.Write(w, binary.BigEndian, g.BoxSize); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, g.BoxType); err != nil {
		return err
	}
	_, err := w.Write(g.Data)
	return err
}

// FtypBox represents the file type box
type FtypBox struct {
	BaseBox
	MajorBrand       [4]byte
	MinorVersion     uint32
	CompatibleBrands [][4]byte
}

// NewFtypBox creates a new ftyp box
func NewFtypBox() *FtypBox {
	return &FtypBox{
		BaseBox: BaseBox{
			BoxType: [4]byte{'f', 't', 'y', 'p'},
		},
		MajorBrand:       [4]byte{'i', 's', 'o', 'm'},
		MinorVersion:     512,
		CompatibleBrands: [][4]byte{{'i', 's', 'o', 'm'}, {'m', 'p', '4', '1'}},
	}
}

// WriteTo writes the ftyp box to a writer
func (f *FtypBox) WriteTo(w io.Writer) error {
	f.BoxSize = uint32(16 + len(f.CompatibleBrands)*4)
	
	if err := binary.Write(w, binary.BigEndian, f.BoxSize); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, f.BoxType); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, f.MajorBrand); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, f.MinorVersion); err != nil {
		return err
	}
	
	for _, brand := range f.CompatibleBrands {
		if err := binary.Write(w, binary.BigEndian, brand); err != nil {
			return err
		}
	}
	
	return nil
}

// ReadFrom reads the ftyp box from a reader
func (f *FtypBox) ReadFrom(r io.Reader) error {
	if err := binary.Read(r, binary.BigEndian, &f.BoxSize); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &f.BoxType); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &f.MajorBrand); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &f.MinorVersion); err != nil {
		return err
	}
	
	// Read compatible brands
	remaining := int(f.BoxSize) - 16
	numBrands := remaining / 4
	f.CompatibleBrands = make([][4]byte, numBrands)
	
	for i := 0; i < numBrands; i++ {
		if err := binary.Read(r, binary.BigEndian, &f.CompatibleBrands[i]); err != nil {
			return err
		}
	}
	
	return nil
}

// MvhdBox represents the movie header box
type MvhdBox struct {
	BaseBox
	Version          uint8
	Flags            [3]byte
	CreationTime     uint32
	ModificationTime uint32
	Timescale        uint32
	Duration         uint32
	Rate             uint32
	Volume           uint16
	Reserved1        uint16
	Reserved2        [2]uint32
	Matrix           [9]uint32
	PreDefined       [6]uint32
	NextTrackID      uint32
}

// NewMvhdBox creates a new mvhd box
func NewMvhdBox(timescale, duration uint32) *MvhdBox {
	now := uint32(time.Now().Unix())
	
	return &MvhdBox{
		BaseBox: BaseBox{
			BoxSize: 108,
			BoxType: [4]byte{'m', 'v', 'h', 'd'},
		},
		Version:          0,
		CreationTime:     now,
		ModificationTime: now,
		Timescale:        timescale,
		Duration:         duration,
		Rate:             0x00010000, // 1.0
		Volume:           0x0100,     // 1.0
		Matrix: [9]uint32{
			0x00010000, 0, 0,
			0, 0x00010000, 0,
			0, 0, 0x40000000,
		},
		NextTrackID: 2,
	}
}

// WriteTo writes the mvhd box to a writer
func (m *MvhdBox) WriteTo(w io.Writer) error {
	if err := binary.Write(w, binary.BigEndian, m.BoxSize); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, m.BoxType); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, m.Version); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, m.Flags); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, m.CreationTime); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, m.ModificationTime); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, m.Timescale); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, m.Duration); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, m.Rate); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, m.Volume); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, m.Reserved1); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, m.Reserved2); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, m.Matrix); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, m.PreDefined); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, m.NextTrackID); err != nil {
		return err
	}
	
	return nil
}

// ReadFrom reads the mvhd box from a reader
func (m *MvhdBox) ReadFrom(r io.Reader) error {
	if err := binary.Read(r, binary.BigEndian, &m.BoxSize); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &m.BoxType); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &m.Version); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &m.Flags); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &m.CreationTime); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &m.ModificationTime); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &m.Timescale); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &m.Duration); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &m.Rate); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &m.Volume); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &m.Reserved1); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &m.Reserved2); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &m.Matrix); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &m.PreDefined); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &m.NextTrackID); err != nil {
		return err
	}
	
	return nil
}

// BoxReader provides functionality to read MP4 boxes from a stream
type BoxReader struct {
	reader io.Reader
}

// NewBoxReader creates a new box reader
func NewBoxReader(r io.Reader) *BoxReader {
	return &BoxReader{reader: r}
}

// ReadBox reads the next box from the stream
func (br *BoxReader) ReadBox() (Box, error) {
	var size uint32
	var boxType [4]byte
	
	if err := binary.Read(br.reader, binary.BigEndian, &size); err != nil {
		return nil, err
	}
	if err := binary.Read(br.reader, binary.BigEndian, &boxType); err != nil {
		return nil, err
	}
	
	// Create appropriate box type
	switch string(boxType[:]) {
	case "ftyp":
		box := &FtypBox{BaseBox: BaseBox{BoxSize: size, BoxType: boxType}}
		// Read the remaining data
		remaining := make([]byte, size-8)
		if _, err := io.ReadFull(br.reader, remaining); err != nil {
			return nil, err
		}
		return box, nil
		
	case "mvhd":
		box := &MvhdBox{BaseBox: BaseBox{BoxSize: size, BoxType: boxType}}
		// Read the remaining data
		remaining := make([]byte, size-8)
		if _, err := io.ReadFull(br.reader, remaining); err != nil {
			return nil, err
		}
		return box, nil
		
	default:
		// Skip unknown boxes  
		remaining := make([]byte, size-8)
		if _, err := io.ReadFull(br.reader, remaining); err != nil {
			return nil, err
		}
		// Create a generic box that implements all interface methods
		box := &GenericBox{BaseBox: BaseBox{BoxSize: size, BoxType: boxType}}
		return box, nil
	}
}

// BoxWriter provides functionality to write MP4 boxes to a stream
type BoxWriter struct {
	writer io.Writer
}

// NewBoxWriter creates a new box writer
func NewBoxWriter(w io.Writer) *BoxWriter {
	return &BoxWriter{writer: w}
}

// WriteBox writes a box to the stream
func (bw *BoxWriter) WriteBox(box Box) error {
	return box.WriteTo(bw.writer)
}

// CreateBasicMP4 creates a basic MP4 file structure
func CreateBasicMP4(w io.Writer, duration time.Duration) error {
	writer := NewBoxWriter(w)
	
	// Write ftyp box
	ftyp := NewFtypBox()
	if err := writer.WriteBox(ftyp); err != nil {
		return fmt.Errorf("failed to write ftyp box: %w", err)
	}
	
	// Write moov box (simplified)
	// In a real implementation, this would include trak boxes and other metadata
	moovSize := uint32(108 + 8) // mvhd + moov header
	if err := binary.Write(w, binary.BigEndian, moovSize); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, [4]byte{'m', 'o', 'o', 'v'}); err != nil {
		return err
	}
	
	// Write mvhd box
	mvhd := NewMvhdBox(90000, uint32(duration.Seconds()*90000))
	if err := writer.WriteBox(mvhd); err != nil {
		return fmt.Errorf("failed to write mvhd box: %w", err)
	}
	
	return nil
}