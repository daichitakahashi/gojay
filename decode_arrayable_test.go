package gojay

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type testArrayable struct {
	Items []*testArrayableObject
	array bool
}

func (a *testArrayable) UnaryJSONObject() UnmarshalerJSONObject {
	a.Items = append(a.Items[:0], &testArrayableObject{})
	// エラー時に配列に追加しない、という選択肢がない… ******************************************************************
	return a.Items[0]
}

func (a *testArrayable) UnmarshalJSONUnary(dec *Decoder) error {
	obj := &testArrayableObject{}
	err := dec.Object(obj)
	if err != nil {
		return err
	}
	a.Items = append(a.Items[:0], obj)
	return nil
}

func (a *testArrayable) UnmarshalJSONArray(dec *Decoder) error {
	obj := &testArrayableObject{}
	err := dec.Object(obj)
	if err != nil {
		return err
	}
	a.Items = append(a.Items, obj)
	return nil
}

type testArrayableObject struct {
	name string
}

func (o *testArrayableObject) UnmarshalJSONObject(dec *Decoder, key string) error {
	if key == "name" {
		return dec.String(&o.name)
	}
	return nil
}

func (o *testArrayableObject) NKeys() int {
	return 1
}

var _ UnmarshalerJSONArrayable = (*testArrayable)(nil)

func TestUnmarshalArrayable(t *testing.T) {
	t.Run("unary", func(t *testing.T) {
		src := strings.NewReader(`{
			"name":"john"
		}`)
		array := &testArrayable{}

		dec := NewDecoder(src)
		err := dec.Decode(array)
		require.NoError(t, err)
		require.Len(t, array.Items, 1)
		require.Equal(t, array.Items[0].name, "john")
	})

	t.Run("array", func(t *testing.T) {
		src := strings.NewReader(`[
			{"name":"john"},
			{"name":"maeda"}
		]`)
		array := &testArrayable{}

		dec := NewDecoder(src)
		err := dec.Decode(array)
		require.NoError(t, err)

		require.Len(t, array.Items, 2)
		require.Equal(t, array.Items[0].name, "john")
		require.Equal(t, array.Items[1].name, "maeda")
	})
}

func TestUnmarshalVariable(t *testing.T) {
	testcases := []struct {
		desc       string
		src        string
		decodeFunc func(t *testing.T, dec *Decoder, c byte)
	}{
		{
			"string",
			`{"value":"samplestring"}`,
			func(t *testing.T, dec *Decoder, c byte) {
				require.Equal(t, byte('"'), c)
				var val string
				err := dec.String(&val)
				require.NoError(t, err)
				require.Equal(t, "samplestring", val)
			},
		}, {
			"int",
			`{"value":1234}`,
			func(t *testing.T, dec *Decoder, c byte) {
				require.Equal(t, byte('1'), c)
				var val int
				err := dec.Int(&val)
				require.NoError(t, err)
				require.Equal(t, 1234, val)
			},
		}, {
			"pseudo-object",
			`{"value":{"mem1":"val1","mem2":true,"mem3":-0.1}}`,
			func(t *testing.T, dec *Decoder, c byte) {
				require.Equal(t, byte('{'), c)
				obj := make(map[string]interface{})
				err := dec.Object(DecodeObjectFunc(func(dec *Decoder, key string) error {
					switch key {
					case "mem1":
						var val string
						err := dec.String(&val)
						require.NoError(t, err)
						obj[key] = val
					case "mem2":
						var val bool
						err := dec.Bool(&val)
						require.NoError(t, err)
						obj[key] = val
					case "mem3":
						var val float64
						err := dec.Float(&val)
						require.NoError(t, err)
						obj[key] = val
					}
					return nil
				}))
				require.NoError(t, err)
				require.Equal(t, "val1", obj["mem1"])
				require.Equal(t, true, obj["mem2"])
				require.Equal(t, -0.1, obj["mem3"])
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.desc, func(t *testing.T) {
			err := Unmarshal(
				[]byte(testcase.src),
				DecodeObjectFunc(func(dec *Decoder, key string) error {
					if key == "value" {
						dec.Variable(DecodeVariableFunc(func(dec *Decoder, c byte) error {
							testcase.decodeFunc(t, dec, c)
							return nil
						}))
						require.Nil(t, dec.err)
					}
					return nil
				}))
			require.NoError(t, err)
		})
	}
}
