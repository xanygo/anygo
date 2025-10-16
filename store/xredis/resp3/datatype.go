//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-29

package resp3

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/big"
	"strconv"
)

var MaxResponseSize = 512 * 1024 * 1024

// DataType 数据类型
// https://redis.io/docs/latest/develop/reference/protocol-spec/
type DataType byte

func (dt DataType) Byte() byte {
	return byte(dt)
}

const (
	DataTypeSimpleString   DataType = '+'
	DataTypeSimpleError    DataType = '-'
	DataTypeInteger        DataType = ':'
	DataTypeBulkString     DataType = '$'
	DataTypeArray          DataType = '*'
	DataTypeNull           DataType = '_'
	DataTypeBoolean        DataType = '#'
	DataTypeDouble         DataType = ','
	DataTypeBigNumber      DataType = '('
	DataTypeBulkError      DataType = '!'
	DataTypeVerbatimString DataType = '='
	DataTypeMap            DataType = '%'
	DataTypeAttribute      DataType = '|'
	DataTypeSet            DataType = '~'
	DataTypePush           DataType = '>'

	DataTypeAny DataType = ' ' // 任意类型，用于构建 Request 数据使用
)

func (dt DataType) Valid() error {
	switch dt {
	case DataTypeSimpleString,
		DataTypeSimpleError,
		DataTypeInteger,
		DataTypeBulkString,
		DataTypeArray,
		DataTypeNull,
		DataTypeBoolean,
		DataTypeDouble,
		DataTypeBigNumber,
		DataTypeBulkError,
		DataTypeVerbatimString,
		DataTypeMap,
		DataTypeAttribute,
		DataTypeSet,
		DataTypePush:
		return nil
	case DataTypeAny:
		return nil
	default:
		return fmt.Errorf("invalid data type: %q", dt)
	}
}

func (dt DataType) IsString() bool {
	return dt == DataTypeSimpleString || dt == DataTypeBulkString
}

func (dt DataType) IsError() bool {
	return dt == DataTypeSimpleError || dt == DataTypeBulkError
}

func (dt DataType) IsArray() bool {
	return dt == DataTypeArray || dt == DataTypeSet
}

func (dt DataType) IsMap() bool {
	return dt == DataTypeMap || dt == DataTypeAttribute
}

func (dt DataType) Equal(b DataType) bool {
	if dt == b || b == DataTypeAny {
		return true
	}
	if dt.IsString() && b.IsString() {
		return true
	} else if dt.IsError() && b.IsError() {
		return true
	} else if dt.IsArray() && b.IsArray() {
		return true
	} else if dt.IsMap() && b.IsMap() {
		return true
	}
	return false
}

func (dt DataType) Load(rd Reader) (Element, error) {
	switch dt {
	case DataTypeSimpleString:
		return dt.loadSimpleString(rd)
	case DataTypeSimpleError:
		return dt.loadSimpleError(rd)
	case DataTypeInteger:
		return dt.loadInteger(rd)
	case DataTypeBulkString:
		return dt.loadBulkString(rd)
	case DataTypeArray:
		return dt.loadArray(rd)
	case DataTypeNull:
		return dt.loadNull(rd)
	case DataTypeBoolean:
		return dt.loadBoolean(rd)
	case DataTypeDouble:
		return dt.loadDouble(rd)
	case DataTypeBigNumber:
		return dt.loadBigNumber(rd)
	case DataTypeBulkError:
		return dt.loadBulkError(rd)
	case DataTypeVerbatimString:
		return dt.loadVerbatimString(rd)
	case DataTypeMap:
		return dt.loadMap(rd)
	case DataTypeAttribute:
		return dt.loadAttribute(rd)
	case DataTypeSet:
		return dt.loadSet(rd)
	case DataTypePush:
		return dt.loadPush(rd)
	default:
		return nil, fmt.Errorf("unkonwn data type: %q", dt)
	}
}

func newInvalidDataError(line []byte) error {
	return fmt.Errorf("%w: %q", ErrInvalidReply, line)
}

func (dt DataType) loadSimpleString(rd Reader) (SimpleString, error) {
	line, err := readLine(rd)
	if err != nil {
		return "", err
	}
	return SimpleString(line), nil
}

func readLine(rd Reader) ([]byte, error) {
	bf, err := rd.ReadSlice('\n')
	if err != nil {
		return nil, err
	}
	if len(bf) < 2 || !bytes.HasSuffix(bf, CRLF) {
		return nil, newInvalidDataError(bf)
	}
	return bf[:len(bf)-2], nil
}

func (dt DataType) loadSimpleError(rd Reader) (SimpleError, error) {
	line, err := readLine(rd)
	if err != nil {
		return "", err
	}
	return SimpleError(line), nil
}

func (dt DataType) loadInteger(rd Reader) (Integer, error) {
	line, err := readLine(rd)
	if err != nil {
		return 0, err
	}
	num, err1 := strconv.ParseInt(string(line), 10, 64)
	if err1 != nil {
		return 0, err1
	}
	return Integer(num), nil
}

func (dt DataType) loadBulkString(rd Reader) (BulkString, error) {
	bf, err := dt.loadBulkBytes(rd)
	if err != nil {
		return "", err
	}
	if bf == nil {
		return "", ErrNil
	}
	return BulkString(bf), nil
}

func (dt DataType) loadBulkBytes(rd Reader) ([]byte, error) {
	line, err := readLine(rd)
	if err != nil {
		return nil, err
	}
	length, err1 := strconv.Atoi(string(line))
	if err1 != nil {
		return nil, err1
	}
	if length < -1 {
		return nil, newInvalidDataError(line)
	}
	if length == -1 {
		// $-1\r\n   -> Null bulk strings
		return nil, nil
	}
	if length > MaxResponseSize {
		return nil, fmt.Errorf("out of MaxResponseSize: %d", length)
	}
	bf := make([]byte, length+2)
	// 将数据以及 \r\n 一起读取出来
	_, err2 := io.ReadFull(rd, bf)
	if err2 != nil {
		return nil, err2
	}
	if !bytes.HasSuffix(bf, CRLF) {
		return nil, newInvalidDataError(bf)
	}
	return bf[:length], nil
}

func (dt DataType) loadArray(rd Reader) (Array, error) {
	return loadArray[Array](rd)
}

type arrayType interface {
	Array | Set | Push
}

func loadArray[T arrayType](rd Reader) (T, error) {
	line, err := readLine(rd)
	if err != nil {
		return nil, err
	}
	length, err1 := strconv.Atoi(string(line))
	if err1 != nil {
		return nil, err1
	}
	if length < -1 {
		return nil, newInvalidDataError(line)
	}
	if length == -1 {
		return nil, ErrNil
	} else if length == 0 {
		return T{}, nil
	}
	result := make(T, 0, length)
	for i := 0; i < length; i++ {
		item, err3 := readOne(rd)
		if err3 != nil {
			// Null elements in arrays
			if errors.Is(err3, ErrNil) {
				result = append(result, nil)
				continue
			}
			return nil, err3
		}
		result = append(result, item)
	}
	return result, nil
}

func (dt DataType) loadNull(rd Reader) (Null, error) {
	line, err := readLine(rd)
	if err != nil {
		return Null{}, err
	}
	if len(line) != 0 {
		return Null{}, newInvalidDataError(line)
	}
	return Null{}, nil
}

func (dt DataType) loadBoolean(rd Reader) (Boolean, error) {
	line, err := readLine(rd)
	if err != nil {
		return false, err
	}
	if len(line) != 1 {
		return false, newInvalidDataError(line)
	}
	switch line[0] {
	case 't':
		return Boolean(true), nil
	case 'f':
		return Boolean(false), nil
	default:
		return false, newInvalidDataError(line)
	}
}

func (dt DataType) loadDouble(rd Reader) (Double, error) {
	line, err := readLine(rd)
	if err != nil {
		return 0, err
	}
	num, err1 := strconv.ParseFloat(string(line), 64)
	if err1 != nil {
		return 0, err1
	}
	return Double(num), nil
}

func (dt DataType) loadBigNumber(rd Reader) (BigNumber, error) {
	line, err := readLine(rd)
	if err != nil {
		return BigNumber{}, err
	}
	bi := new(big.Int)
	if _, ok := bi.SetString(string(line), 10); !ok {
		return BigNumber{}, newInvalidDataError(line)
	}
	return BigNumber(*bi), nil
}

func (dt DataType) loadBulkError(rd Reader) (BulkError, error) {
	bf, err := dt.loadBulkBytes(rd)
	if err != nil {
		return "", err
	}
	return BulkError(bf), nil
}

func (dt DataType) loadVerbatimString(rd Reader) (VerbatimString, error) {
	bf, err := dt.loadBulkBytes(rd)
	if err != nil {
		return VerbatimString{}, err
	}
	enc, data, found := bytes.Cut(bf, []byte(":"))
	if !found {
		if len(bf) > 100 {
			bf = bf[:100]
		}
		return VerbatimString{}, newInvalidDataError(bf)
	}
	return VerbatimString{
		Encoding: string(enc),
		Data:     string(data),
	}, nil
}

func (dt DataType) loadMap(rd Reader) (Map, error) {
	return loadMap[Map](rd)
}

type mapsType interface {
	Map | Attribute
}

func loadMap[T mapsType](rd Reader) (T, error) {
	line, err := readLine(rd)
	if err != nil {
		return nil, err
	}
	num, err := strconv.Atoi(string(line))
	if err != nil {
		return nil, err
	}
	result := make(T, num)
	for i := 0; i < num; i++ {
		// read key:
		key, err3 := readOne(rd)
		if err3 != nil {
			return nil, err3
		}
		// read value:
		value, err5 := readOne(rd)
		if err5 != nil {
			return nil, err5
		}
		result[key] = value
	}
	return result, nil
}

func readOne(rd Reader) (Element, error) {
	dt, err := rd.ReadByte()
	if err != nil {
		return nil, err
	}
	return DataType(dt).Load(rd)
}

func (dt DataType) loadAttribute(rd Reader) (Attribute, error) {
	return loadMap[Attribute](rd)
}

func (dt DataType) loadSet(rd Reader) (Set, error) {
	return loadArray[Set](rd)
}

func (dt DataType) loadPush(rd Reader) (Push, error) {
	return loadArray[Push](rd)
}
