/*
Copyright 2014 Tamás Gulácsi

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

package mnb

import (
	"bytes"
	"fmt"
	"time"

	"gopkg.in/inf.v0"
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

type Double inf.Dec

func (d Double) String() string {
	return (*inf.Dec)(&d).String()
}

func (d Double) MarshalText() ([]byte, error) {
	return (*inf.Dec)(&d).MarshalText()
}

func (d *Double) UnmarshalText(data []byte) error {
	i := bytes.IndexByte(data, ',')
	if i >= 0 {
		data[i] = '.'
	}
	if _, ok := (*inf.Dec)(d).SetString(string(data)); !ok {
		return fmt.Errorf("error parsing %q", data)
	}
	return nil
}
