package query

import (
	"testing"

	"github.com/Jeffail/benthos/v3/lib/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMethods(t *testing.T) {
	type easyMsg struct {
		content string
		meta    map[string]string
	}

	tests := map[string]struct {
		input    string
		output   string
		messages []easyMsg
		index    int
	}{
		"literal function": {
			input:    `5.from(0)`,
			output:   `5`,
			messages: []easyMsg{{}},
		},
		"json from all": {
			input:  `json("foo").from_all()`,
			output: `["a","b","c"]`,
			messages: []easyMsg{
				{content: `{"foo":"a"}`},
				{content: `{"foo":"b"}`},
				{content: `{"foo":"c"}`},
			},
		},
		"json from all 2": {
			input:  `json("foo").from_all()`,
			output: `["a",null,"c",null]`,
			messages: []easyMsg{
				{content: `{"foo":"a"}`},
				{content: `{}`},
				{content: `{"foo":"c"}`},
				{content: `not even json`},
			},
		},
		"json from all/or": {
			input:  `json("foo").or("fallback").from_all()`,
			output: `["a","fallback","c","fallback"]`,
			messages: []easyMsg{
				{content: `{"foo":"a"}`},
				{content: `{}`},
				{content: `{"foo":"c"}`},
				{content: `not even json`},
			},
		},
		"json from all/or 2": {
			input:  `(json().foo | "fallback").from_all()`,
			output: `["a","fallback","c","fallback"]`,
			messages: []easyMsg{
				{content: `{"foo":"a"}`},
				{content: `{}`},
				{content: `{"foo":"c"}`},
				{content: `not even json`},
			},
		},
		"json from all/or 3": {
			input:  `json().foo.or("fallback").from_all()`,
			output: `["a","fallback","c","fallback"]`,
			messages: []easyMsg{
				{content: `{"foo":"a"}`},
				{content: `{}`},
				{content: `{"foo":"c"}`},
				{content: `not even json`},
			},
		},
		"deleted to or": {
			input:    `deleted().or("fallback")`,
			output:   `fallback`,
			messages: []easyMsg{{}},
		},
		"nothing to or": {
			input:    `nothing().or("fallback")`,
			output:   `fallback`,
			messages: []easyMsg{{}},
		},
		"json catch": {
			input:  `json().catch("nope")`,
			output: `nope`,
			messages: []easyMsg{
				{content: `this %$#% isnt json`},
			},
		},
		"json catch 2": {
			input:  `json().catch("nope")`,
			output: `null`,
			messages: []easyMsg{
				{content: `null`},
			},
		},
		"json catch 3": {
			input:  `json("foo").catch("nope")`,
			output: `null`,
			messages: []easyMsg{
				{content: `{"foo":null}`},
			},
		},
		"json catch 4": {
			input:  `json("foo").catch("nope")`,
			output: `yep`,
			messages: []easyMsg{
				{content: `{"foo":"yep"}`},
			},
		},
		"meta from all": {
			input:  `meta("foo").from_all()`,
			output: `["bar","","baz"]`,
			messages: []easyMsg{
				{meta: map[string]string{"foo": "bar"}},
				{},
				{meta: map[string]string{"foo": "baz"}},
			},
		},
		"or json null": {
			input:  `json("foo").or("backup")`,
			output: `backup`,
			messages: []easyMsg{
				{content: `{"foo":null}`},
			},
		},
		"or json null 2": {
			input:  `json("foo").or("backup")`,
			output: `backup`,
			messages: []easyMsg{
				{content: `{"bar":"nope"}`},
			},
		},
		"or json null 3": {
			input:  `json("foo").or(json("bar"))`,
			output: `yep`,
			messages: []easyMsg{
				{content: `{"bar":"yep"}`},
			},
		},
		"or boolean from all": {
			input:  `json("foo").or( json("bar") == "yep" ).from_all()`,
			output: `["from foo",true,false,"from foo 2"]`,
			messages: []easyMsg{
				{content: `{"foo":"from foo"}`},
				{content: `{"bar":"yep"}`},
				{content: `{"bar":"nope"}`},
				{content: `{"foo":"from foo 2","bar":"yep"}`},
			},
		},
		"or boolean from metadata": {
			input:  `meta("foo").or( meta("bar") == "yep" ).from_all()`,
			output: `["from foo",true,false,"from foo 2"]`,
			messages: []easyMsg{
				{meta: map[string]string{"foo": "from foo"}},
				{meta: map[string]string{"bar": "yep"}},
				{meta: map[string]string{"bar": "nope"}},
				{meta: map[string]string{"foo": "from foo 2", "bar": "yep"}},
			},
		},
		"for each": {
			input:  `json("foo").for_each(this + 10)`,
			output: `[11,12,12]`,
			messages: []easyMsg{
				{content: `{"foo":[1,2,2]}`},
			},
		},
		"for each inner map": {
			input:  `json("foo").for_each((this.bar + 10) | "woops")`,
			output: `[11,"woops",12]`,
			messages: []easyMsg{
				{content: `{"foo":[{"bar":1},2,{"bar":2}]}`},
			},
		},
		"for each some errors": {
			input:  `json("foo").for_each((this + 10) | "failed")`,
			output: `[11,12,"failed",12]`,
			messages: []easyMsg{
				{content: `{"foo":[1,2,"nope",2]}`},
			},
		},
		"for each uncaught errors": {
			input:  `json("foo").for_each(this + 10)`,
			output: `[11,12,10,12]`,
			messages: []easyMsg{
				{content: `{"foo":[1,2,"nope",2]}`},
			},
		},
		"for each delete some elements": {
			input: `json("foo").for_each(
	match this {
		this < 10 => deleted()
		_ => this - 10
	}
)`,
			output: `[1,2,3]`,
			messages: []easyMsg{
				{content: `{"foo":[11,12,7,13]}`},
			},
		},
		"for each delete all elements for some reason": {
			input:  `json("foo").for_each(deleted())`,
			output: `[]`,
			messages: []easyMsg{
				{content: `{"foo":[11,12,7,13]}`},
			},
		},
		"for each not an array": {
			input:  `json("foo").for_each(this + 10)`,
			output: `not an array`,
			messages: []easyMsg{
				{content: `{"foo":"not an array"}`},
			},
		},
		"test sum standard array": {
			input:  `json("foo").sum()`,
			output: `5`,
			messages: []easyMsg{
				{content: `{"foo":[1,2,2]}`},
			},
		},
		"test sum standard array 2": {
			input:  `json("foo").sum()`,
			output: `8`,
			messages: []easyMsg{
				{content: `{"foo":[1,2,2,"nah",3]}`},
			},
		},
		"test sum standard array 3": {
			input:  `json("foo").sum()`,
			output: `12`,
			messages: []easyMsg{
				{content: `{"foo":[1,2,2,"4",3]}`},
			},
		},
		"test sum standard array 4": {
			input:  `json("foo").from_all().sum()`,
			output: `16`,
			messages: []easyMsg{
				{content: `{"foo":1}`},
				{content: `{"foo":3}`},
				{content: `{"foo":4}`},
				{content: `{"foo":8}`},
			},
		},
		"test sum standard array 5": {
			input:  `json("foo").from_all().sum()`,
			output: `16`,
			messages: []easyMsg{
				{content: `{"foo":1}`},
				{content: `{"foo":"3"}`},
				{content: `{"foo":"nope"}`},
				{content: `{"foo":4}`},
				{content: `{"foo":8}`},
			},
		},
		"test map json": {
			input:  `json("foo").map(bar)`,
			output: `yep`,
			messages: []easyMsg{
				{content: `{"foo":{"bar":"yep"}}`},
			},
		},
		"test map json 2": {
			input:  `json("foo").map(bar + 10)`,
			output: `13`,
			messages: []easyMsg{
				{content: `{"foo":{"bar":"3"}}`},
			},
		},
		"test map json 3": {
			input:  `json("foo").map(("static"))`,
			output: `static`,
			messages: []easyMsg{
				{content: `{"foo":{"bar":"3"}}`},
			},
		},
		"test string method": {
			input:    `5.string() == "5"`,
			output:   `true`,
			messages: []easyMsg{{}},
		},
		"test number method": {
			input:    `"5".number() == 5`,
			output:   `true`,
			messages: []easyMsg{{}},
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			msg := message.New(nil)
			for _, m := range test.messages {
				part := message.NewPart([]byte(m.content))
				if m.meta != nil {
					for k, v := range m.meta {
						part.Metadata().Set(k, v)
					}
				}
				msg.Append(part)
			}

			e, err := tryParse(test.input, false)
			require.NoError(t, err)
			res := e.ToString(FunctionContext{
				Index: test.index,
				Msg:   msg,
			})
			assert.Equal(t, test.output, res)
			res = string(e.ToBytes(FunctionContext{
				Index: test.index,
				Msg:   msg,
			}))
			assert.Equal(t, test.output, res)
		})
	}
}

func TestMethodErrors(t *testing.T) {
	type easyMsg struct {
		content string
		meta    map[string]string
	}

	tests := map[string]struct {
		input    string
		errStr   string
		messages []easyMsg
		index    int
	}{
		"literal function": {
			input:    `"not a number".number()`,
			errStr:   "strconv.ParseFloat: parsing \"not a number\": invalid syntax",
			messages: []easyMsg{{}},
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			msg := message.New(nil)
			for _, m := range test.messages {
				part := message.NewPart([]byte(m.content))
				if m.meta != nil {
					for k, v := range m.meta {
						part.Metadata().Set(k, v)
					}
				}
				msg.Append(part)
			}

			e, err := tryParse(test.input, false)
			require.NoError(t, err)

			_, err = e.Exec(FunctionContext{
				Index: test.index,
				Msg:   msg,
			})
			assert.EqualError(t, err, test.errStr)
		})
	}
}

func TestMethodMaps(t *testing.T) {
	type easyMsg struct {
		content string
		meta    map[string]string
	}

	tests := map[string]struct {
		input    string
		output   interface{}
		err      string
		maps     map[string]Function
		messages []easyMsg
		index    int
	}{
		"no maps": {
			input:    `"foo".apply("nope")`,
			err:      "no maps were found",
			messages: []easyMsg{{}},
		},
		"map not exist": {
			input:    `"foo".apply("nope")`,
			err:      "map nope was not found",
			maps:     map[string]Function{},
			messages: []easyMsg{{}},
		},
		"map static": {
			input:  `"foo".apply("foo")`,
			output: "hello world",
			maps: map[string]Function{
				"foo": literalFunction("hello world"),
			},
			messages: []easyMsg{{}},
		},
		"map context": {
			input:  `json().apply("foo")`,
			output: "this value",
			maps: map[string]Function{
				"foo": func() Function {
					f, _ := fieldFunction("foo")
					return f
				}(),
			},
			messages: []easyMsg{{
				content: `{"foo":"this value"}`,
			}},
		},
		"map dynamic": {
			input:  `json().apply(meta("dyn_map"))`,
			output: "this value",
			maps: map[string]Function{
				"foo": func() Function {
					f, _ := fieldFunction("foo")
					return f
				}(),
				"bar": func() Function {
					f, _ := fieldFunction("bar")
					return f
				}(),
			},
			messages: []easyMsg{{
				content: `{"foo":"this value","bar":"and this value"}`,
				meta: map[string]string{
					"dyn_map": "foo",
				},
			}},
		},
		"map dynamic 2": {
			input:  `json().apply(meta("dyn_map"))`,
			output: "and this value",
			maps: map[string]Function{
				"foo": func() Function {
					f, _ := fieldFunction("foo")
					return f
				}(),
				"bar": func() Function {
					f, _ := fieldFunction("bar")
					return f
				}(),
			},
			messages: []easyMsg{{
				content: `{"foo":"this value","bar":"and this value"}`,
				meta: map[string]string{
					"dyn_map": "bar",
				},
			}},
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			msg := message.New(nil)
			for _, m := range test.messages {
				part := message.NewPart([]byte(m.content))
				if m.meta != nil {
					for k, v := range m.meta {
						part.Metadata().Set(k, v)
					}
				}
				msg.Append(part)
			}

			e, err := tryParse(test.input, false)
			require.NoError(t, err)

			res, err := e.Exec(FunctionContext{
				Maps:  test.maps,
				Index: test.index,
				Msg:   msg,
			})
			if len(test.err) > 0 {
				require.EqualError(t, err, test.err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, test.output, res)
		})
	}
}