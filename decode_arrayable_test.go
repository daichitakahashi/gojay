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
