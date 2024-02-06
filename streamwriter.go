// Copyright 2024 Dave van Soest. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package johanson provides streaming JSON utilities.
package johanson

import (
	"encoding/json"
	"io"
	"strconv"
)

type jsonContext interface {
	prewrite(w io.Writer) bool
	pause(on bool)
}

type jsonContextArray struct {
	paused   *bool
	nonEmpty bool
}

func (ctx *jsonContextArray) prewrite(w io.Writer) bool {
	if ctx.nonEmpty {
		w.Write([]byte{','})
	}
	ctx.nonEmpty = true
	return false // Multi value context.
}

func (ctx *jsonContextArray) pause(on bool) {
	*ctx.paused = on
}

type val struct {
	w      io.Writer
	ctx    jsonContext
	paused bool
}

// The type used for the JSON value context.
// Don't assume that this type is a pointer or not, because this
// may change in the future.
type V = *val

type jsonContextSingle struct {
	paused *bool
}

func (jsonContextSingle) prewrite(w io.Writer) bool {
	return true // Single value context.
}

func (ctx jsonContextSingle) pause(on bool) {
	*ctx.paused = on
}

func (v *val) prewrite() (ok bool, single bool) {
	if v == nil || v.w == nil || v.ctx == nil || v.paused {
		return false, false
	} else {
		single := v.ctx.prewrite(v.w)
		return true, single
	}
}

func (v *val) postwrite(single bool) {
	if v.ctx != nil {
		v.ctx.pause(false)
		if single {
			v.ctx = nil
		}
	}
}

// Write `null` to the stream.
// Only works if the value context is in a valid state.
func (v *val) Null() {
	if ok, single := v.prewrite(); ok {
		v.w.Write([]byte("null"))
		v.postwrite(single)
	}
}

// Marshal any value of any type.
// Returns an error if and only if marshaling fails, in which case
// nothing is written to the stream.
// Only works if the value context is in a valid state.
func (v *val) Marshal(value interface{}) error {
	if ok, single := v.prewrite(); ok {
		bytes, err := json.Marshal(value)
		if err == nil {
			v.w.Write(bytes)
		}
		v.postwrite(single)
		if err != nil {
			return err
		}
	}
	return nil
}

// Write either `true` or `false` to the stream.
// Only works if the value context is in a valid state.
func (v *val) Bool(val bool) {
	if ok, single := v.prewrite(); ok {
		if val {
			v.w.Write([]byte("true"))
		} else {
			v.w.Write([]byte("false"))
		}
		v.postwrite(single)
	}
}

// Write the provided signed integer value to the stream.
// Only works if the value context is in a valid state.
func (v *val) Int(val int64) {
	if ok, single := v.prewrite(); ok {
		v.w.Write([]byte(strconv.FormatInt(val, 10)))
		v.postwrite(single)
	}
}

// Write the provided unsigned integer value to the stream.
// Only works if the value context is in a valid state.
func (v *val) Uint(val uint64) {
	if ok, single := v.prewrite(); ok {
		v.w.Write([]byte(strconv.FormatUint(val, 10)))
		v.postwrite(single)
	}
}

// Write the provided floating point value to the stream.
// Only works if the value context is in a valid state.
func (v *val) Float(value float64) {
	if ok, single := v.prewrite(); ok {
		bytes, _ := json.Marshal(value)
		v.w.Write(bytes)
		v.postwrite(single)
	}
}

// Write the provided string value to the stream, escaping special JSON
// characters.
// Only works if the value context is in a valid state.
func (v *val) String(s string) {
	if ok, single := v.prewrite(); ok {
		bytes, _ := json.Marshal(s)
		v.w.Write(bytes)
		v.postwrite(single)
	}
}

// Opens an array context and calls the callback with a value context, from
// which values can be written to the array.
// The array is closed when the callback returns.
func (v *val) Array(fn func(V)) {
	if ok, single := v.prewrite(); ok {
		v.ctx.pause(true)
		v.w.Write([]byte{'['})
		if fn != nil {
			a := val{w: v.w}
			ctx := jsonContextArray{paused: &a.paused}
			a.ctx = &ctx
			fn(&a)
			a.w = nil
		}
		v.w.Write([]byte{']'})
		v.postwrite(single)
	}
}

type obj struct {
	w        io.Writer
	nonEmpty bool
	paused   bool
}

// The type used for the JSON object context.
// Don't assume that this type is a pointer or not, because this
// may change in the future.
type K = *obj

func (o *obj) prewrite() bool {
	if o.w == nil && o.paused {
		return false // Write not allowed.
	} else {
		if o.nonEmpty {
			o.w.Write([]byte{','})
		}
		o.nonEmpty = true
		return true // Write allowed.
	}
}

type jsonContextObjectItem struct {
	key string
	obj *obj
}

func (ctx *jsonContextObjectItem) pause(on bool) {
	ctx.obj.paused = on
}

func (ctx *jsonContextObjectItem) prewrite(w io.Writer) bool {
	if ctx.obj.prewrite() {
		bytes, _ := json.Marshal(ctx.key)
		w.Write(bytes)
		w.Write([]byte{':'})
	}
	return true // Single value context.
}

// Prepare for adding an item to the JSON object context.
// The item will use key as its key.
// Returns a JSON single value context, to which the value of the item can be
// written.
// Only works if the object context is in a valid state.
func (o *obj) Item(key string) V {
	if !o.paused {
		o.paused = true
		return &val{w: o.w, ctx: &jsonContextObjectItem{key: key, obj: o}}
	}
	return nil
}

// Marshal any object/map and write its contents to the JSON object context.
// Only works if the object context is in a valid state.
func (o *obj) Marshal(anyMap map[string]interface{}) error {
	bytes, err := json.Marshal(anyMap)
	if err != nil {
		return err
	}
	if len(bytes) > 2 && o.prewrite() {
		o.w.Write(bytes[1 : len(bytes)-1])
	}
	return nil
}

// Opens an object context and calls the callback with the object context, from
// which items can be written to the object.
// The object is closed when the callback returns.
func (v *val) Object(fn func(K)) {
	if ok, single := v.prewrite(); ok {
		v.ctx.pause(true)
		v.w.Write([]byte{'{'})
		if fn != nil {
			jso := obj{w: v.w}
			fn(&jso)
			jso.w = nil
		}
		v.w.Write([]byte{'}'})
		v.postwrite(single)
	}
}

type writerWrapper struct {
	w   io.Writer
	Err error
}

// Implements the io.Writer interface.
// Keeps track of the last occurred error.
func (ww *writerWrapper) Write(p []byte) (n int, err error) {
	n, err = ww.w.Write(p)
	if err != nil {
		ww.Err = err
	}
	return
}

// Check whether the JSON value context is finished.
func (v *val) Finished() bool {
	return v.ctx == nil
}

// Return the last error that occurred while writing to the stream, or `nil`
// in case no error has occurred.
func (v *val) Error() error {
	ww, ok := v.w.(*writerWrapper)
	if ok {
		return ww.Err
	} else {
		return nil
	}
}

// NewStreamWriter instantiates a new JSON stream writer, using w as the
// underlying writer.
// It returns a JSON single value context to one value can be written.
func NewStreamWriter(w io.Writer) V {
	ww := &writerWrapper{w: w}
	v := &val{w: ww}
	v.ctx = jsonContextSingle{paused: &v.paused}
	return v
}
