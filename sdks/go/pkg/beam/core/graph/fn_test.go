// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package graph

import (
	"context"
	"reflect"
	"testing"
)

func TestNewCombineFn(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		tests := []struct {
			cfn interface{}
		}{
			{cfn: func(int, int) int { return 0 }},
			{cfn: func(string, string) string { return "" }},
			{cfn: func(MyAccum, MyAccum) MyAccum { return MyAccum{} }},
			{cfn: func(MyAccum, MyAccum) (MyAccum, error) { return MyAccum{}, nil }},
			{cfn: func(context.Context, MyAccum, MyAccum) MyAccum { return MyAccum{} }},
			{cfn: func(context.Context, MyAccum, MyAccum) (MyAccum, error) { return MyAccum{}, nil }},
			{cfn: &GoodCombineFn{}},
			{cfn: &GoodWErrorCombineFn{}},
			{cfn: &GoodWContextCombineFn{}},
			{cfn: &GoodCombineFnUnexportedExtraMethod{}},
		}

		for _, test := range tests {
			t.Run(reflect.TypeOf(test.cfn).String(), func(t *testing.T) {
				if _, err := NewCombineFn(test.cfn); err != nil {
					t.Fatalf("NewCombineFn failed: %v", err)
				}
			})
		}
	})
	t.Run("invalid", func(t *testing.T) {
		tests := []struct {
			cfn interface{}
		}{
			// Validate MergeAccumulator errors
			{cfn: func() int { return 0 }},
			{cfn: func(int, int) {}},
			{cfn: func(int, int) string { return "" }},
			{cfn: func(string, string) int { return 0 }},
			{cfn: func(int, string) int { return 0 }},
			{cfn: func(string, int) int { return 0 }},
			{cfn: func(string, int) (int, error) { return 0, nil }},
			{cfn: &BadCombineFnNoMergeAccumulators{}},
			{cfn: &BadCombineFnNonBinaryMergeAccumulators{}},
			// Validate accumulator type mismatches
			{cfn: &BadCombineFnMisMatchedCreateAccumulator{}},
			{cfn: &BadCombineFnMisMatchedAddInputIn{}},
			{cfn: &BadCombineFnMisMatchedAddInputOut{}},
			{cfn: &BadCombineFnMisMatchedAddInputBoth{}},
			{cfn: &BadCombineFnMisMatchedExtractOutput{}},
			// Validate signatures
			{cfn: &BadCombineFnInvalidCreateAccumulator1{}},
			{cfn: &BadCombineFnInvalidCreateAccumulator2{}},
			{cfn: &BadCombineFnInvalidCreateAccumulator3{}},
			{cfn: &BadCombineFnInvalidCreateAccumulator4{}},
			{cfn: &BadCombineFnInvalidAddInput1{}},
			{cfn: &BadCombineFnInvalidAddInput2{}},
			{cfn: &BadCombineFnInvalidAddInput3{}},
			{cfn: &BadCombineFnInvalidAddInput4{}},
			{cfn: &BadCombineFnInvalidExtractOutput1{}},
			{cfn: &BadCombineFnInvalidExtractOutput2{}},
			{cfn: &BadCombineFnInvalidExtractOutput3{}},
			{cfn: &BadCombineFnExtraExportedMethod{}},
		}
		for _, test := range tests {
			t.Run(reflect.TypeOf(test.cfn).String(), func(t *testing.T) {
				if cfn, err := NewCombineFn(test.cfn); err != nil {
					// Note to Developer: To work on improving the error messages, use t.Errorf instead!
					t.Logf("NewCombineFn failed as expected:\n%v", err)
				} else {
					t.Errorf("AsCombineFn(%v) = %v, want failure", cfn.Name(), cfn)
				}
			})
		}
	})
}

// Do not copy. The following types are for testing signatures only.
// They are not working examples.
// Keep all test functions Above this point.
type MyAccum struct{}

// Examples of correct CombineFn signatures

type GoodCombineFn struct{}

func (fn *GoodCombineFn) MergeAccumulators(MyAccum, MyAccum) MyAccum {
	return MyAccum{}
}

func (fn *GoodCombineFn) CreateAccumulator() MyAccum {
	return MyAccum{}
}

func (fn *GoodCombineFn) AddInput(MyAccum, int) MyAccum {
	return MyAccum{}
}

func (fn *GoodCombineFn) ExtractOutput(MyAccum) int64 {
	return 0
}

type GoodWErrorCombineFn struct{}

func (fn *GoodWErrorCombineFn) MergeAccumulators(int, int) (int, error) {
	return 0, nil
}

type GoodWContextCombineFn struct{}

func (fn *GoodWContextCombineFn) MergeAccumulators(context.Context, MyAccum, MyAccum) MyAccum {
	return MyAccum{}
}

func (fn *GoodWContextCombineFn) CreateAccumulator(context.Context) MyAccum {
	return MyAccum{}
}

func (fn *GoodWContextCombineFn) AddInput(context.Context, MyAccum, int) MyAccum {
	return MyAccum{}
}

func (fn *GoodWContextCombineFn) ExtractOutput(context.Context, MyAccum) int64 {
	return 0
}

type GoodCombineFnUnexportedExtraMethod struct {
	*GoodCombineFn
}

func (fn *GoodCombineFnUnexportedExtraMethod) unexportedExtraMethod(context.Context, string) string {
	return ""
}

// Examples of incorrect CombineFn signatures.
// Embedding *GoodCombineFn avoids repetitive MergeAccumulators signatures when desired.
// The immeadiately following examples are relating to accumulator mismatches.

type BadCombineFnNoMergeAccumulators struct{}

func (fn *BadCombineFnNoMergeAccumulators) CreateAccumulator() string { return "" }

type BadCombineFnNonBinaryMergeAccumulators struct {
	*GoodCombineFn
}

func (fn *BadCombineFnNonBinaryMergeAccumulators) MergeAccumulators(int, string) int {
	return 0
}

type BadCombineFnMisMatchedCreateAccumulator struct {
	*GoodCombineFn
}

func (fn *BadCombineFnMisMatchedCreateAccumulator) CreateAccumulator() string {
	return ""
}

type BadCombineFnMisMatchedAddInputIn struct {
	*GoodCombineFn
}

func (fn *BadCombineFnMisMatchedAddInputIn) AddInput(string, int) MyAccum {
	return MyAccum{}
}

type BadCombineFnMisMatchedAddInputOut struct {
	*GoodCombineFn
}

func (fn *BadCombineFnMisMatchedAddInputOut) AddInput(MyAccum, int) string {
	return ""
}

type BadCombineFnMisMatchedAddInputBoth struct {
	*GoodCombineFn
}

func (fn *BadCombineFnMisMatchedAddInputBoth) AddInput(string, int) string {
	return ""
}

type BadCombineFnMisMatchedExtractOutput struct {
	*GoodCombineFn
}

func (fn *BadCombineFnMisMatchedExtractOutput) ExtractOutput(string) int {
	return 0
}

// Examples of incorrect CreateAccumulator signatures

type BadCombineFnInvalidCreateAccumulator1 struct {
	*GoodCombineFn
}

func (fn *BadCombineFnInvalidCreateAccumulator1) CreateAccumulator(context.Context, string) int {
	return 0
}

type BadCombineFnInvalidCreateAccumulator2 struct {
	*GoodCombineFn
}

func (fn *BadCombineFnInvalidCreateAccumulator2) CreateAccumulator(string) int {
	return 0
}

type BadCombineFnInvalidCreateAccumulator3 struct {
	*GoodCombineFn
}

func (fn *BadCombineFnInvalidCreateAccumulator3) CreateAccumulator() (MyAccum, string) {
	return MyAccum{}, ""
}

type BadCombineFnInvalidCreateAccumulator4 struct {
	*GoodCombineFn
}

func (fn *BadCombineFnInvalidCreateAccumulator4) CreateAccumulator() (string, MyAccum) {
	return "", MyAccum{}
}

// Examples of incorrect AddInput signatures

type BadCombineFnInvalidAddInput1 struct {
	*GoodCombineFn
}

func (fn *BadCombineFnInvalidAddInput1) AddInput(context.Context, string) int {
	return 0
}

type BadCombineFnInvalidAddInput2 struct {
	*GoodCombineFn
}

func (fn *BadCombineFnInvalidAddInput2) AddInput(string) int {
	return 0
}

type BadCombineFnInvalidAddInput3 struct {
	*GoodCombineFn
}

func (fn *BadCombineFnInvalidAddInput3) AddInput(context.Context, string, string, string) int {
	return 0
}

type BadCombineFnInvalidAddInput4 struct {
	*GoodCombineFn
}

func (fn *BadCombineFnInvalidAddInput4) AddInput(MyAccum, string) (int, int, int) {
	return 0, 0, 0
}

// Examples of incorrect ExtractOutput signatures

type BadCombineFnInvalidExtractOutput1 struct {
	*GoodCombineFn
}

func (fn *BadCombineFnInvalidExtractOutput1) ExtractOutput(MyAccum, string) (int, int, int) {
	return 0, 0, 0
}

type BadCombineFnInvalidExtractOutput2 struct {
	*GoodCombineFn
}

func (fn *BadCombineFnInvalidExtractOutput2) ExtractOutput() (int, int, int) {
	return 0, 0, 0
}

type BadCombineFnInvalidExtractOutput3 struct {
	*GoodCombineFn
}

func (fn *BadCombineFnInvalidExtractOutput3) ExtractOutput(context.Context, MyAccum, int) int {
	return 0
}

// Other CombineFn Errors

type BadCombineFnExtraExportedMethod struct {
	*GoodCombineFn
}

func (fn *BadCombineFnExtraExportedMethod) ExtraMethod(string) int {
	return 0
}
