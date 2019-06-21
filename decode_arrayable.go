package gojay

// UnmarshalerJSONVariable is
type UnmarshalerJSONVariable interface {
	UnmarshalJSONVariable(*Decoder, byte) error
}

// DecodeVariable is
func (dec *Decoder) DecodeVariable(v UnmarshalerJSONVariable) error {
	if dec.isPooled == 1 {
		panic(InvalidUsagePooledDecoderError("Invalid usage of pooled decoder"))
	}
	return dec.decodeVariable(v)
}

func (dec *Decoder) decodeVariable(v UnmarshalerJSONVariable) error {
	for ; dec.cursor < dec.length || dec.read(); dec.cursor++ {
		switch dec.data[dec.cursor] {
		case ' ', '\n', '\t', '\r', ',':
			continue
		case 'n':
			// is null
			dec.cursor++
			err := dec.assertNull()
			if err != nil {
				return err
			}
			return nil
		case '{', '[', '"', 'f', 't', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-':
			// dec.cursor++
			err := v.UnmarshalJSONVariable(dec, dec.data[dec.cursor])
			if err != nil {
				dec.err = err
				return dec.err
			}
			return nil
		default:
			return dec.raiseInvalidJSONErr(dec.cursor)
		}
	}
	return dec.raiseInvalidJSONErr(dec.cursor)
}

// DecodeVariableFunc is a func type implementing UnmarshalerJSONVariable.
type DecodeVariableFunc func(*Decoder, byte) error

// UnmarshalJSONVariable implements UnmarshalerJSONVariable.
func (f DecodeVariableFunc) UnmarshalJSONVariable(dec *Decoder, c byte) error {
	return f(dec, c)
}

// Add Values functions

// AddVariable decodes the JSON value within an object or an array to a UnmarshalerJSONVariable.
func (dec *Decoder) AddVariable(v UnmarshalerJSONVariable) error {
	return dec.Variable(v)
}

// Variable decodes the JSON value within an object or an array to a UnmarshalerJSONVariable.
func (dec *Decoder) Variable(v UnmarshalerJSONVariable) error {
	return dec.decodeVariable(v)
}

// UnmarshalerJSONArrayable is
type UnmarshalerJSONArrayable interface {
	UnmarshalerJSONArray
	UnmarshalJSONUnary(*Decoder) error
}

// DecodeArrayable is
func (dec *Decoder) DecodeArrayable(v UnmarshalerJSONArrayable) error {
	if dec.isPooled == 1 {
		panic(InvalidUsagePooledDecoderError("Invalid usage of pooled decoder"))
	}
	return dec.decodeArrayable(v, false)
}
func (dec *Decoder) decodeArrayable(arr UnmarshalerJSONArrayable, insideObj bool) error {
	for ; dec.cursor < dec.length || dec.read(); dec.cursor++ {
		switch dec.data[dec.cursor] {
		case ' ', '\n', '\t', '\r', ',':
			continue
		case '[':
			var err error
			if insideObj {
				err = dec.Array(arr)
			} else {
				_, err = dec.decodeArray(arr)
			}
			return err
		case 'n':
			// is null
			dec.cursor++
			err := dec.assertNull()
			if err != nil {
				return err
			}
			return nil
		case '{':
			err := arr.UnmarshalJSONUnary(dec)
			return err
		case '"', 'f', 't', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			// can't unmarshall to array or struct
			// we skip and set Error
			dec.err = dec.makeInvalidUnmarshalErr(arr)
			err := dec.skipData()
			if err != nil {
				return err
			}
			return nil
		default:
			return dec.raiseInvalidJSONErr(dec.cursor)
		}
	}
	return dec.raiseInvalidJSONErr(dec.cursor)
}

func (dec *Decoder) decodeArrayableNull(v interface{}) error {
	for ; dec.cursor < dec.length || dec.read(); dec.cursor++ {
		switch dec.data[dec.cursor] {
		case ' ', '\n', '\t', '\r', ',':
			continue
		case '[':
			return dec.ArrayNull(v)
		case 'n':
			// is null
			dec.cursor++
			err := dec.assertNull()
			if err != nil {
				return err
			}
			return nil
		case '{':
			arr, ok := v.(UnmarshalerJSONArrayable)
			if !ok {
				dec.err = dec.makeInvalidUnmarshalErr((UnmarshalerJSONArrayable)(nil))
				return dec.err
			}
			return arr.UnmarshalJSONUnary(dec)
		case '"', 'f', 't', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			// can't unmarshall to array or struct
			// we skip and set Error
			dec.err = dec.makeInvalidUnmarshalErr((UnmarshalerJSONArrayable)(nil))
			err := dec.skipData()
			if err != nil {
				return err
			}
			return nil
		default:
			return dec.raiseInvalidJSONErr(dec.cursor)
		}
	}
	return dec.raiseInvalidJSONErr(dec.cursor)
}

// DecodeArrayableFunc is a func type implementing UnmarshalerJSONArrayable.
// Use it to cast a `func(*Decoder) error` to Unmarshal an array on the fly.
type DecodeArrayableFunc func(*Decoder) error

// UnmarshalJSONArrayable implements UnmarshalerJSONArray.
func (f DecodeArrayableFunc) UnmarshalJSONArrayable(dec *Decoder) error {
	return f(dec)
}

// IsNil implements UnmarshalerJSONArrayable.
func (f DecodeArrayableFunc) IsNil() bool {
	return f == nil
}

// Add Values functions

// AddArrayable decodes the JSON value within an object or an array to a UnmarshalerJSONArrayable.
func (dec *Decoder) AddArrayable(v UnmarshalerJSONArrayable) error {
	return dec.Arrayable(v)
}

// AddArrayableNull decodes the JSON value within an object or an array to a UnmarshalerJSONArrayable.
func (dec *Decoder) AddArrayableNull(v interface{}) error {
	return dec.ArrayableNull(v)
}

// Arrayable decodes the JSON value within an object or an array to a UnmarshalerJSONArrayable.
func (dec *Decoder) Arrayable(v UnmarshalerJSONArrayable) error {
	return dec.decodeArrayable(v, true)
}

// ArrayableNull decodes the JSON value within an object or an array to a UnmarshalerJSONArrayable.
// v should be a pointer to an UnmarshalerJSONArrayable,
// if `null` value is encountered in JSON, it will leave the value v untouched,
// else it will create a new instance of the UnmarshalerJSONArrayable behind v.
func (dec *Decoder) ArrayableNull(v interface{}) error {
	return dec.decodeArrayableNull(v)
}
