/*
Copyright 2019, 2021 Tamás Gulácsi

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
//
// SPDX-License-Identifier: Apache-2.0

package mnb

import (
	"bytes"
	"fmt"
	"time"

	"github.com/cockroachdb/apd/v3"
)

type Date time.Time

func (d Date) String() string {
	return time.Time(d).Format("2006-01-02")
}

func (d Date) MarshalText() ([]byte, error) {
	return []byte(time.Time(d).Format("2006-01-02")), nil
}

func (d *Date) UnmarshalText(data []byte) error {
	t, err := time.Parse("2006-01-02", string(data))
	if err != nil {
		return err
	}
	*d = Date(t)
	return nil
}

type Double struct {
	*apd.Decimal
}

func NewDouble(coeff int64, exponent int32) Double {
	return Double{Decimal: apd.New(coeff, exponent)}
}
func NewDoubleFromString(s string) (Double, error) {
	var d apd.Decimal
	err := d.Scan(s)
	return Double{Decimal: &d}, err
}

func (d Double) String() string {
	return d.Decimal.Text('f')
}

func (d Double) MarshalText() ([]byte, error) {
	return []byte(d.Decimal.Text('f')), nil
}

func (d *Double) UnmarshalText(data []byte) error {
	i := bytes.IndexByte(data, ',')
	if i >= 0 {
		data[i] = '.'
	}
	var err error
	if d.Decimal == nil {
		d.Decimal, _, err = apd.NewFromString(string(data))
	} else {
		_, _, err = d.Decimal.SetString(string(data))
	}
	if err != nil {
		return fmt.Errorf("%q: %w", string(data), err)
	}
	return nil
}
