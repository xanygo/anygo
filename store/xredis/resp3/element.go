//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-29

package resp3

import (
	"bytes"
	"encoding"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"

	"github.com/xanygo/anygo/ds/xsync"
)

var CRLF = []byte{'\r', '\n'}

type Element interface {
	Bytes(bf *bytes.Buffer) []byte
	DataType() DataType
}

var _ Element = SimpleString("")

// SimpleString are encoded as a plus (+) character, followed by a string.
// The string mustn't contain a CR (\r) or LF (\n) character and is terminated by CRLF (i.e., \r\n).
type SimpleString string

func (s SimpleString) Bytes(bf *bytes.Buffer) []byte {
	// +OK\r\n
	bf.Reset()
	bf.WriteByte(DataTypeSimpleString.Byte())
	bf.WriteString(string(s))
	bf.Write(CRLF)
	return bf.Bytes()
}

func (s SimpleString) DataType() DataType {
	return DataTypeSimpleString
}

func (s SimpleString) String() string {
	return string(s)
}

func (s SimpleString) ToFloat64() (float64, error) {
	return strconv.ParseFloat(string(s), 64)
}

func (s SimpleString) ToInt64() (int64, error) {
	return strconv.ParseInt(string(s), 10, 64)
}

func (s SimpleString) ToUint64() (uint64, error) {
	return strconv.ParseUint(string(s), 10, 64)
}

var _ Element = Integer(0)

type Integer int64

func (i Integer) Bytes(bf *bytes.Buffer) []byte {
	// :[<+|->]<value>\r\n
	bf.Reset()
	bf.WriteByte(DataTypeInteger.Byte())
	return bf.Bytes()
}

func (i Integer) Int() int {
	return int(i)
}

func (i Integer) Int64() int64 {
	return int64(i)
}

func (i Integer) DataType() DataType {
	return DataTypeInteger
}

var _ Element = BulkString("")

// BulkString A bulk string represents a single binary string.
// The string can be of any size, but by default,
// Redis limits it to 512 MB (see the proto-max-bulk-len configuration directive).
type BulkString string

func (b BulkString) Bytes(bf *bytes.Buffer) []byte {
	// $<length>\r\n<data>\r\n
	bf.Reset()
	bf.WriteByte(DataTypeBulkString.Byte())
	bf.WriteString(strconv.Itoa(len(b)))
	bf.Write(CRLF)
	bf.WriteString(string(b))
	bf.Write(CRLF)
	return bf.Bytes()
}

func (b BulkString) DataType() DataType {
	return DataTypeBulkString
}

func (b BulkString) String() string {
	return string(b)
}

func (b BulkString) ToFloat64() (float64, error) {
	return strconv.ParseFloat(string(b), 64)
}

func (b BulkString) ToInt64() (int64, error) {
	return strconv.ParseInt(string(b), 10, 64)
}

func (b BulkString) ToUint64() (uint64, error) {
	return strconv.ParseUint(string(b), 10, 64)
}

var _ Element = Array(nil)

type Array []Element

var bp = xsync.NewBytesBufferPool(1024)

func (a Array) Bytes(bf *bytes.Buffer) []byte {
	// *<number-of-elements>\r\n<element-1>...<element-n>
	bf.Reset()
	bf.WriteByte(DataTypeArray.Byte())
	bf.WriteString(strconv.Itoa(len(a)))
	bf.Write(CRLF)
	b := bp.Get()
	for _, e := range a {
		bf.Write(e.Bytes(b))
	}
	bp.Put(b)
	return bf.Bytes()
}

func (a Array) DataType() DataType {
	return DataTypeArray
}

// ToZSlice 返回 Sorted zet 的结果列表
func (a Array) ToZSlice() ([]Z, error) {
	if len(a)%2 != 0 {
		return nil, fmt.Errorf("expected even number of keys, got %d", len(a))
	}
	result := make([]Z, 0, len(a)/2)
	for i := 0; i < len(a); i += 2 {
		member, err1 := ToString(a[i], nil)
		if err1 != nil {
			return nil, err1
		}
		score, err2 := ToFloat64(a[i+1], nil)
		if err2 != nil {
			return nil, err2
		}
		item := Z{
			Member: member,
			Score:  score,
		}
		result = append(result, item)
	}
	return result, nil
}

var _ Element = Null{}

type Null struct{}

func (n Null) Bytes(bf *bytes.Buffer) []byte {
	return []byte("_\r\n")
}

func (n Null) DataType() DataType {
	return DataTypeNull
}

var _ Element = Boolean(true)

type Boolean bool

func (b Boolean) Bytes(bf *bytes.Buffer) []byte {
	if b {
		return []byte("#t\r\n")
	}
	return []byte("#f\r\n")
}

func (b Boolean) DataType() DataType {
	return DataTypeBoolean
}

func (b Boolean) Bool() bool {
	return bool(b)
}

var _ Element = Double(1)

type Double float64

func (b Double) Bytes(bf *bytes.Buffer) []byte {
	// ,[<+|->]<integral>[.<fractional>][<E|e>[sign]<exponent>]\r\n
	bf.Reset()
	bf.WriteByte(DataTypeDouble.Byte())
	bf.WriteString(strconv.FormatFloat(float64(b), 'g', -1, 64))
	bf.Write(CRLF)
	return bf.Bytes()
}

func (b Double) DataType() DataType {
	return DataTypeDouble
}

func (b Double) Float64() float64 {
	return float64(b)
}

var _ Element = BigNumber(big.Int{})

// BigNumber This type can encode integer values outside the range of signed 64-bit integers.
type BigNumber big.Int

func (bn BigNumber) Bytes(bf *bytes.Buffer) []byte {
	// ([+|-]<number>\r\n
	bf.Reset()
	bf.WriteByte(DataTypeBigNumber.Byte())
	bi := big.Int(bn)
	bf.WriteString((&bi).String())
	return bf.Bytes()
}

func (bn BigNumber) DataType() DataType {
	return DataTypeBigNumber
}

func (bn BigNumber) BigInt() *big.Int {
	return (*big.Int)(&bn)
}

func (bn BigNumber) Int64() int64 {
	return bn.BigInt().Int64()
}

var _ Element = VerbatimString{}

type VerbatimString struct {
	Encoding string // 默认为 txt
	Data     string
}

func (vs VerbatimString) getEncoding() string {
	if vs.Encoding == "" {
		return "txt"
	}
	return vs.Encoding
}

func (vs VerbatimString) DataType() DataType {
	return DataTypeVerbatimString
}

func (vs VerbatimString) Bytes(bf *bytes.Buffer) []byte {
	// =<length>\r\n<encoding>:<data>\r\n
	bf.Reset()

	encoding := vs.getEncoding()

	bf.WriteByte(DataTypeVerbatimString.Byte())
	bf.WriteString(strconv.Itoa(len(encoding) + 1 + len(vs.Data)))
	bf.WriteString(encoding)
	bf.WriteByte(':')
	bf.WriteString(vs.Data)
	bf.Write(CRLF)
	return bf.Bytes()
}

var _ Element = Map(nil)
var _ encoding.TextMarshaler = Map(nil)
var _ json.Marshaler = Map(nil)

// Map The RESP map encodes a collection of key-value tuples, i.e., a dictionary or a hash.
//
// Both map keys and values can be any of RESP's types.
type Map map[Element]Element

func (m Map) Bytes(bf *bytes.Buffer) []byte {
	// %<number-of-entries>\r\n<key-1><value-1>...<key-n><value-n>
	bf.Reset()
	bf.WriteByte(DataTypeMap.Byte())
	bf.WriteString(strconv.Itoa(len(m)))
	bf.Write(CRLF)
	b := bp.Get()
	for k, v := range m {
		bf.Write(k.Bytes(b))
		bf.Write(CRLF)
		bf.Write(v.Bytes(b))
	}
	bp.Put(b)
	return bf.Bytes()
}

func (m Map) DataType() DataType {
	return DataTypeMap
}

func (m Map) MarshalJSON() ([]byte, error) {
	mp, err := mapToStringAnyMap(m)
	if err != nil {
		return nil, err
	}
	return json.Marshal(mp)
}

func (m Map) MarshalText() ([]byte, error) {
	bf := bp.Get()
	text := m.Bytes(bf)
	bp.Put(bf)
	return text, nil
}

func (m Map) ToStringMap() (map[string]string, error) {
	return mapToStringMap(m)
}

func (m Map) ToStringAnyMap() (map[string]any, error) {
	return mapToStringAnyMap(m)
}

var _ Element = Attribute(nil)

// Attribute The attribute type is exactly like the Map type, but instead of a % character as the first byte,
// the | character is used.
// Attributes describe a dictionary exactly like the Map type.
// However the client should not consider such a dictionary part of the reply,
// but as auxiliary data that augments the reply.
type Attribute map[Element]Element

func (ab Attribute) Bytes(bf *bytes.Buffer) []byte {
	bf.Reset()
	bf.WriteByte(DataTypeAttribute.Byte())
	bf.WriteString(strconv.Itoa(len(ab)))
	bf.Write(CRLF)
	b := bp.Get()
	for k, v := range ab {
		bf.Write(k.Bytes(b))
		bf.Write(CRLF)
		bf.Write(v.Bytes(b))
	}
	bp.Put(b)
	return bf.Bytes()
}

func (ab Attribute) DataType() DataType {
	return DataTypeAttribute
}

func (ab Attribute) ToStringMap() (map[string]string, error) {
	return mapToStringMap(ab)
}

func (ab Attribute) MarshalJSON() ([]byte, error) {
	mp, err := mapToStringAnyMap(ab)
	if err != nil {
		return nil, err
	}
	return json.Marshal(mp)
}

func (ab Attribute) MarshalText() ([]byte, error) {
	bf := bp.Get()
	text := ab.Bytes(bf)
	bp.Put(bf)
	return text, nil
}

func (ab Attribute) ToStringAnyMap() (map[string]any, error) {
	return mapToStringAnyMap(ab)
}

var _ Element = Set(nil)

// Set  are somewhat like Arrays but are unordered and should only contain unique elements.
type Set []Element

func (s Set) Bytes(bf *bytes.Buffer) []byte {
	// ~<number-of-elements>\r\n<element-1>...<element-n>
	bf.Reset()
	bf.WriteByte(DataTypeSet.Byte())
	bf.WriteString(strconv.Itoa(len(s)))
	bf.Write(CRLF)
	b := bp.Get()
	for _, e := range s {
		bf.Write(e.Bytes(b))
	}
	bp.Put(b)
	return bf.Bytes()
}

func (s Set) DataType() DataType {
	return DataTypeSet
}

var _ Element = Push(nil)

type Push []Element

func (p Push) Bytes(bf *bytes.Buffer) []byte {
	// ><number-of-elements>\r\n<element-1>...<element-n>
	bf.Reset()
	bf.WriteByte(DataTypePush.Byte())
	bf.WriteString(strconv.Itoa(len(p)))
	bf.Write(CRLF)
	b := bp.Get()
	for _, e := range p {
		bf.Write(e.Bytes(b))
	}
	bp.Put(b)
	return bf.Bytes()
}

func (p Push) DataType() DataType {
	return DataTypePush
}
