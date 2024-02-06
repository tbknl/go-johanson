package johanson_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/tbknl/go-johanson"
)

func Test_BasicTypes(t *testing.T) {
	testCases := []struct {
		expected string
		fn       func(johanson.V)
	}{
		{expected: `null`, fn: func(v johanson.V) { v.Null() }},
		{expected: `true`, fn: func(v johanson.V) { v.Bool(true) }},
		{expected: `false`, fn: func(v johanson.V) { v.Bool(false) }},
		{expected: `-123`, fn: func(v johanson.V) { v.Int(-123) }},
		{expected: `456`, fn: func(v johanson.V) { v.Uint(456) }},
		{expected: `987.654`, fn: func(v johanson.V) { v.Float(987.654) }},
		{expected: `"abc"`, fn: func(v johanson.V) { v.String("abc") }},
	}

	for i, tc := range testCases {
		w := &strings.Builder{}
		v := johanson.NewStreamWriter(w)
		tc.fn(v)
		if want, got := tc.expected, w.String(); want != got {
			t.Errorf("Test case %d: got %s instead of %s", i, got, want)
		}
	}
}

func Test_StringEscaping(t *testing.T) {
	testCases := []struct {
		value string
	}{
		{"tab \t tab"},
		{"newline \n newline"},
		{"return \r return"},
		{"backslash \\ backslash"},
		{`double quotes " double quotes`},
		{"character \u0018 below 0x20"},
		{"tab \t newline \n return \n backslash \\ double quotes \" \u0018 the end"},
	}

	for i, tc := range testCases {
		w := &strings.Builder{}
		v := johanson.NewStreamWriter(w)
		v.String(tc.value)
		want, _ := json.Marshal(tc.value)
		if got := w.String(); string(want) != got {
			t.Errorf("String escaping test case %d: got %s instead of %s", i, got, want)
		}
	}
}

func Test_StringObjectKeyEscaping(t *testing.T) {
	testCases := []struct {
		key string
	}{
		{"tab \t tab"},
		{"newline \n newline"},
		{"return \r return"},
		{"backslash \\ backslash"},
		{`double quotes " double quotes`},
		{"character \u0018 below 0x20"},
		{"tab \t newline \n return \n backslash \\ double quotes \" \u0018 the end"},
	}

	for i, tc := range testCases {
		w := &strings.Builder{}
		v := johanson.NewStreamWriter(w)
		v.Object(func(obj johanson.K) {
			obj.Item(tc.key).Int(int64(i))
		})
		want, _ := json.Marshal(map[string]int{tc.key: i})
		if got := w.String(); string(want) != got {
			t.Errorf("Test case %d: got %s instead of %s", i, got, want)
		}
	}
}

func Test_IgnoreWriteDataAfterFinish(t *testing.T) {
	w := &strings.Builder{}
	v := johanson.NewStreamWriter(w)
	v.Int(123)
	v.String("Ignored")
	if want, got := "123", w.String(); want != got {
		t.Fatalf("Ignore write after finish failed: got %s instead of %s", got, want)
	}
}

func Test_Array(t *testing.T) {
	testCases := []struct {
		expected string
		fn       func(johanson.V)
	}{
		{expected: `[]`, fn: func(v johanson.V) { v.Array(nil) }},
		{
			expected: `[123]`,
			fn: func(v johanson.V) {
				v.Array(func(a johanson.V) {
					a.Int(123)
				})
			},
		},
		{
			expected: `[123,"abc",["nested"],true]`,
			fn: func(v johanson.V) {
				v.Array(func(a johanson.V) {
					a.Int(123)
					a.String("abc")
					a.Array(func(a2 johanson.V) {
						a2.String("nested")
					})
					a.Bool(true)
				})
			},
		},
	}

	for i, tc := range testCases {
		w := &strings.Builder{}
		v := johanson.NewStreamWriter(w)
		tc.fn(v)
		if want, got := tc.expected, w.String(); want != got {
			t.Errorf("Test case %d: got %s instead of %s", i, got, want)
		}
	}
}

func Test_IgnoreArrayWriteInNestedContext(t *testing.T) {
	w := &strings.Builder{}
	v := johanson.NewStreamWriter(w)

	v.Array(func(a johanson.V) {
		a.Int(1)
		a.Object(func(o johanson.K) {
			a.Int(2) // Ignored!
		})
		a.Int(3)
	})

	if want, got := "[1,{},3]", w.String(); want != got {
		t.Errorf("Test case: got %s instead of %s", got, want)
	}
}

func Test_IgnoreParentWriteInArrayContext(t *testing.T) {
	w := &strings.Builder{}
	v := johanson.NewStreamWriter(w)

	v.Array(func(a johanson.V) {
		v.Int(123)
		a.Int(456)
		v.Int(789)
	})

	if want, got := `[456]`, w.String(); want != got {
		t.Errorf("Test case: got %s instead of %s", got, want)
	}
}

func Test_Object(t *testing.T) {
	testCases := []struct {
		expected string
		fn       func(johanson.V)
	}{
		{expected: `{}`, fn: func(v johanson.V) { v.Object(nil) }},
		{
			expected: `{"x":123}`,
			fn: func(v johanson.V) {
				v.Object(func(obj johanson.K) {
					obj.Item("x").Int(123)
				})
			},
		},
		{
			expected: `{"x":123,"nested":{"y":[false]},"z":{}}`,
			fn: func(v johanson.V) {
				v.Object(func(obj johanson.K) {
					obj.Item("x").Int(123)
					obj.Item("nested").Object(func(nested johanson.K) {
						nested.Item("y").Array(func(a johanson.V) {
							a.Bool(false)
						})
					})
					obj.Item("z").Object(nil)
				})
			},
		},
	}

	for i, tc := range testCases {
		w := &strings.Builder{}
		v := johanson.NewStreamWriter(w)
		tc.fn(v)
		if want, got := tc.expected, w.String(); want != got {
			t.Errorf("Test case %d: got %s instead of %s", i, got, want)
		}
	}
}

func Test_MarshalInsideObject(t *testing.T) {
	testCases := []struct {
		expected string
		fn       func(johanson.V)
	}{
		{expected: `{}`, fn: func(v johanson.V) { v.Marshal(map[string]int{}) }},
		{
			expected: `{"x":1,"y":"Z"}`,
			fn: func(v johanson.V) {
				v.Object(func(obj johanson.K) {
					obj.Marshal(map[string]interface{}{
						"x": 1,
						"y": "Z",
					})
				})
			},
		},
		{
			expected: `{"y":null}`,
			fn: func(v johanson.V) {
				v.Object(func(obj johanson.K) {
					obj.Marshal(map[string]interface{}{})
					obj.Item("y").Null()
				})
			},
		},
		{
			expected: `{"mX":1,"y":null}`,
			fn: func(v johanson.V) {
				v.Object(func(obj johanson.K) {
					obj.Marshal(map[string]interface{}{
						"mX": 1,
					})
					obj.Item("y").Null()
				})
			},
		},
		{
			expected: `{"x":123,"mA":"abc","mB":"def","y":true}`,
			fn: func(v johanson.V) {
				v.Object(func(obj johanson.K) {
					obj.Item("x").Int(123)
					obj.Marshal(map[string]interface{}{
						"mA": "abc",
						"mB": "def",
					})
					obj.Item("y").Bool(true)
				})
			},
		},
	}

	for i, tc := range testCases {
		w := &strings.Builder{}
		v := johanson.NewStreamWriter(w)
		tc.fn(v)
		if want, got := tc.expected, w.String(); want != got {
			t.Errorf("Test case %d: got %s instead of %s", i, got, want)
		}
	}
}

func Test_ObjectItemsWithoutValue(t *testing.T) {
	testCases := []struct {
		expected string
		fn       func(johanson.V)
	}{
		{
			expected: `{}`,
			fn: func(v johanson.V) {
				v.Object(func(obj johanson.K) {
					obj.Item("ignored")
				})
			},
		},
		{
			expected: `{"x":123}`,
			fn: func(v johanson.V) {
				v.Object(func(obj johanson.K) {
					obj.Item("x").Int(123)
					obj.Item("ignored")
				})
			},
		},
		{
			expected: `{}`,
			fn: func(v johanson.V) {
				v.Object(func(obj johanson.K) {
					obj.Item("ignored")
					obj.Item("x").Int(123)
				})
			},
		},
		{
			expected: `{"x":123}`,
			fn: func(v johanson.V) {
				v.Object(func(obj johanson.K) {
					obj.Item("x").Int(123)
					obj.Item("ignored")
					obj.Item("y").Int(456)
				})
			},
		},
	}

	for i, tc := range testCases {
		w := &strings.Builder{}
		v := johanson.NewStreamWriter(w)
		tc.fn(v)
		if want, got := tc.expected, w.String(); want != got {
			t.Errorf("Test case %d: got %s instead of %s", i, got, want)
		}
	}
}

func Test_IgnoreObjectWriteInNestedContext(t *testing.T) {
	w := &strings.Builder{}
	v := johanson.NewStreamWriter(w)

	v.Object(func(o johanson.K) {
		o.Item("a").Int(1)
		o.Item("b").Array(func(_ johanson.V) {
			o.Item("c").Int(2)
		})
		o.Item("d").Object(func(_ johanson.K) {
			o.Item("e").Int(3)
		})
		o.Item("f").Int(4)
	})

	if want, got := `{"a":1,"b":[],"d":{},"f":4}`, w.String(); want != got {
		t.Errorf("Test case: got %s instead of %s", got, want)
	}
}

func Test_IgnoreParentWriteInObjectContext(t *testing.T) {
	w := &strings.Builder{}
	v := johanson.NewStreamWriter(w)

	v.Object(func(o johanson.K) {
		v.Int(123)
		o.Item("a").Int(456)
		v.Int(789)
	})

	if want, got := `{"a":456}`, w.String(); want != got {
		t.Errorf("Test case: got %s instead of %s", got, want)
	}
}

func Test_IgnoreObjectItemSubsequentValues(t *testing.T) {
	testCases := []struct {
		expected string
		fn       func(johanson.V)
	}{
		{
			expected: `{"x":123}`,
			fn: func(v johanson.V) {
				v.Object(func(obj johanson.K) {
					item := obj.Item("x")
					item.Int(123)
					item.Int(456)
				})
			},
		},
		{
			expected: `{"x":123,"y":456}`,
			fn: func(v johanson.V) {
				v.Object(func(obj johanson.K) {
					item := obj.Item("x")
					item.Int(123)
					item.String("Ignored!")
					item.Int(789)
					obj.Item("y").Int(456)
				})
			},
		},
	}

	for i, tc := range testCases {
		w := &strings.Builder{}
		v := johanson.NewStreamWriter(w)
		tc.fn(v)
		if want, got := tc.expected, w.String(); want != got {
			t.Errorf("Test case %d: got %s instead of %s", i, got, want)
		}
	}
}

func Test_Marshal(t *testing.T) {
	testCases := []struct {
		expected string
		fn       func(johanson.V)
	}{
		{
			expected: `{"x":1,"y":"Z"}`,
			fn: func(v johanson.V) {
				v.Marshal(map[string]interface{}{"x": 1, "y": "Z"})
			},
		},
		{
			expected: `[123,"abc"]`,
			fn: func(v johanson.V) {
				v.Marshal([]interface{}{123, "abc"})
			},
		},
	}

	for i, tc := range testCases {
		w := &strings.Builder{}
		v := johanson.NewStreamWriter(w)
		tc.fn(v)
		if want, got := tc.expected, w.String(); want != got {
			t.Errorf("Test case %d: got %s instead of %s", i, got, want)
		}
	}
}

func Test_MarshalError(t *testing.T) {
	{
		w := &strings.Builder{}
		v := johanson.NewStreamWriter(w)
		err := v.Marshal("No error expected!")
		if want, got := error(nil), err; want != got {
			t.Errorf("Marshal error: got %s instead of %s", got, want)
		}
	}

	{
		w := &strings.Builder{}
		v := johanson.NewStreamWriter(w)
		err := v.Marshal(func() {}) // NOTE: Functions can't be marshaled.
		if err == nil {
			t.Errorf("Marshal error expected: got nil instead")
		}
	}
}

func Test_Finished(t *testing.T) {
	w := &strings.Builder{}
	v := johanson.NewStreamWriter(w)

	if want, got := false, v.Finished(); want != got {
		t.Fatalf("Expected new stream to not be finished but instead it is.")
	}

	v.Array(func(a johanson.V) {
		a.Int(1)

		if want, got := false, v.Finished(); want != got {
			t.Fatalf("Expected stream with open array to not be finished but instead it is.")
		}
	})

	if want, got := true, v.Finished(); want != got {
		t.Fatalf("Expected stream to be finished but instead it is not.")
	}
}

type StringBuilderWrapper struct {
	strings.Builder
	Limit int
}

func (sbw *StringBuilderWrapper) Write(p []byte) (int, error) {
	n, err := sbw.Builder.Write(p)
	if err == nil {
		if sbw.Len() > sbw.Limit {
			err = fmt.Errorf("Buffer length %d exceeds limit %d.", sbw.Len(), sbw.Limit)
		}
	}
	return n, err
}

func Test_WriterErrorCheck(t *testing.T) {
	w := &StringBuilderWrapper{Limit: 10}
	v := johanson.NewStreamWriter(w)

	if v.Error() != nil {
		t.Fatalf("Expected new stream to not have writer error but instead it has.")
	}

	v.Array(func(a johanson.V) {
		a.Int(12345)

		if v.Error() != nil {
			t.Fatalf("Expected stream to not have writer error but instead it has.")
		}

		a.Int(67890)

		if v.Error() == nil {
			t.Fatalf("Expected stream to have writer error but instead it has not.")
		}
	})

	if v.Error() == nil {
		t.Fatalf("Expected stream to have writer error but instead it has not.")
	}
}
