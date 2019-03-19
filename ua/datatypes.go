// Copyright 2018-2019 opcua authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package ua

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// These flags define which fields of a DataValue are set.
// Bits are or'ed together if multiple fields are set.
const (
	DataValueValue             = 0x1
	DataValueStatus            = 0x2
	DataValueSourceTimestamp   = 0x4
	DataValueServerTimestamp   = 0x8
	DataValueSourcePicoseconds = 0x10
	DataValueServerPicoseconds = 0x20
)

// DataValue is always preceded by a mask that indicates which fields are present in the stream.
//
// Specification: Part 6, 5.2.2.17
type DataValue struct {
	EncodingMask      byte
	Value             *Variant
	Status            uint32
	SourceTimestamp   time.Time
	SourcePicoseconds uint16
	ServerTimestamp   time.Time
	ServerPicoseconds uint16
}

func (d *DataValue) Decode(b []byte) (int, error) {
	buf := NewBuffer(b)
	d.EncodingMask = buf.ReadByte()
	if d.Has(DataValueValue) {
		d.Value = new(Variant)
		buf.ReadStruct(d.Value)
	}
	if d.Has(DataValueStatus) {
		d.Status = buf.ReadUint32()
	}
	if d.Has(DataValueSourceTimestamp) {
		d.SourceTimestamp = buf.ReadTime()
	}
	if d.Has(DataValueSourcePicoseconds) {
		d.SourcePicoseconds = buf.ReadUint16()
	}
	if d.Has(DataValueServerTimestamp) {
		d.ServerTimestamp = buf.ReadTime()
	}
	if d.Has(DataValueServerPicoseconds) {
		d.ServerPicoseconds = buf.ReadUint16()
	}
	return buf.Pos(), buf.Error()
}

func (d *DataValue) Encode() ([]byte, error) {
	buf := NewBuffer(nil)
	buf.WriteUint8(d.EncodingMask)

	if d.Has(DataValueValue) {
		buf.WriteStruct(d.Value)
	}
	if d.Has(DataValueStatus) {
		buf.WriteUint32(d.Status)
	}
	if d.Has(DataValueSourceTimestamp) {
		buf.WriteTime(d.SourceTimestamp)
	}
	if d.Has(DataValueSourcePicoseconds) {
		buf.WriteUint16(d.SourcePicoseconds)
	}
	if d.Has(DataValueServerTimestamp) {
		buf.WriteTime(d.ServerTimestamp)
	}
	if d.Has(DataValueServerPicoseconds) {
		buf.WriteUint16(d.ServerPicoseconds)
	}
	return buf.Bytes(), buf.Error()
}

func (d *DataValue) Has(mask byte) bool {
	return d.EncodingMask&mask == mask
}

func (d *DataValue) UpdateMask() {
	d.EncodingMask = 0
	if d.Value != nil {
		d.EncodingMask |= DataValueValue
	}
	if d.Status != 0 {
		d.EncodingMask |= DataValueStatus
	}
	if !d.SourceTimestamp.IsZero() {
		d.EncodingMask |= DataValueSourceTimestamp
	}
	if !d.ServerTimestamp.IsZero() {
		d.EncodingMask |= DataValueServerTimestamp
	}
	if d.SourcePicoseconds > 0 {
		d.EncodingMask |= DataValueSourcePicoseconds
	}
	if d.ServerPicoseconds > 0 {
		d.EncodingMask |= DataValueServerPicoseconds
	}
}

// GUID represents GUID in binary stream. It is a 16-byte globally unique identifier.
//
// Specification: Part 6, 5.1.3
type GUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 uint64
}

// NewGUID creates a new GUID.
// Input should be GUID string of 16 hexadecimal characters like 1111AAAA-22BB-33CC-44DD-55EE77FF9900.
// Dash can be omitted, and alphabets are not case-sensitive.
func NewGUID(guid string) *GUID {
	h := strings.Replace(guid, "-", "", -1)
	b, err := hex.DecodeString(h)
	if err != nil {
		return nil
	}
	if len(b) < 16 {
		return nil
	}

	return &GUID{
		Data1: binary.BigEndian.Uint32(b[:4]),
		Data2: binary.BigEndian.Uint16(b[4:6]),
		Data3: binary.BigEndian.Uint16(b[6:8]),
		Data4: binary.BigEndian.Uint64(b[8:16]),
	}
}

func (g *GUID) Decode(b []byte) (int, error) {
	buf := NewBuffer(b)
	g.Data1 = buf.ReadUint32()
	g.Data2 = buf.ReadUint16()
	g.Data3 = buf.ReadUint16()
	g.Data4 = buf.ReadUint64()
	return buf.Pos(), buf.Error()
}

func (g *GUID) Encode() ([]byte, error) {
	buf := NewBuffer(nil)
	buf.WriteUint32(g.Data1)
	buf.WriteUint16(g.Data2)
	buf.WriteUint16(g.Data3)
	buf.WriteUint64(g.Data4)
	return buf.Bytes(), buf.Error()
}

// String returns GUID in human-readable string.
func (g *GUID) String() string {
	d4 := make([]byte, 8)
	binary.BigEndian.PutUint64(d4, g.Data4)

	return fmt.Sprintf("%X-%X-%X-%X-%X",
		g.Data1,
		g.Data2,
		g.Data3,
		d4[:2],
		d4[2:],
	)
}

// These flags define which fields of a LocalizedText are set.
// Bits are or'ed together if multiple fields are set.
const (
	LocalizedTextLocale = 0x1
	LocalizedTextText   = 0x2
)

// LocalizedText represents a LocalizedText.
// A LocalizedText structure contains two fields that could be missing.
// For that reason, the encoding uses a bit mask to indicate which fields
// are actually present in the encoded forl.
//
// Specification: Part 6, 5.2.2.14
type LocalizedText struct {
	EncodingMask uint8
	Locale       string
	Text         string
}

func (l *LocalizedText) Decode(b []byte) (int, error) {
	buf := NewBuffer(b)
	l.EncodingMask = buf.ReadByte()
	l.Locale = ""
	l.Text = ""
	if l.Has(LocalizedTextLocale) {
		l.Locale = buf.ReadString()
	}
	if l.Has(LocalizedTextText) {
		l.Text = buf.ReadString()
	}
	return buf.Pos(), buf.Error()
}

func (l *LocalizedText) Encode() ([]byte, error) {
	buf := NewBuffer(nil)
	buf.WriteUint8(l.EncodingMask)
	if l.Has(LocalizedTextLocale) {
		buf.WriteString(l.Locale)
	}
	if l.Has(LocalizedTextText) {
		buf.WriteString(l.Text)
	}
	return buf.Bytes(), buf.Error()
}

func (l *LocalizedText) Has(mask byte) bool {
	return l.EncodingMask&mask == mask
}

func (l *LocalizedText) UpdateMask() {
	l.EncodingMask = 0
	if l.Locale != "" {
		l.EncodingMask |= LocalizedTextLocale
	}
	if l.Text != "" {
		l.EncodingMask |= LocalizedTextText
	}
}

type StatusCode uint32

type XmlElement string
