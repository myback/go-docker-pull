/*
Copyright © 2021 myback.space <git@myback.space>

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

package progressbar

import (
	"fmt"
	"math"
	"os"
	"sync/atomic"
)

type ProgressBar struct {
	width         int8
	printLenLine  int32
	cur           uint64
	contentLength uint64
	description   string
}

func (pb *ProgressBar) ContentLength(l int64) {
	atomic.SwapUint64(&pb.cur, 0)
	atomic.SwapUint64(&pb.contentLength, uint64(l))
}

func (pb *ProgressBar) SetDescription(s string) {
	pb.description = s
}

func (pb *ProgressBar) Close() {
	fmt.Fprintln(os.Stdout)
}

func (pb *ProgressBar) Flush() {
	desc := pb.description
	for i := 0; i < int(pb.printLenLine); i++ {
		desc += " "
	}

	atomic.SwapInt32(&pb.printLenLine, 0)
	fmt.Fprintf(os.Stdout, "\r%s", desc)
}

func (pb *ProgressBar) fill() string {
	var fill int
	if pb.contentLength == 0 {
		fill = int(pb.width)
	} else {
		fill = int(float64(pb.width) * float64(pb.cur) / float64(pb.contentLength))
	}

	if fill > 0 && pb.contentLength != 0 {
		fill--
	}

	var pbFill string
	for i := 0; i < fill; i++ {
		pbFill += "="
	}

	if pb.contentLength != 0 {
		pbFill += ">"
	}

	return pbFill
}

func (pb *ProgressBar) Write(p []byte) (n int, err error) {
	n = len(p)
	atomic.AddUint64(&pb.cur, uint64(n))

	l, err := fmt.Fprintf(os.Stdout, "\r%s[%-50s] %7s/%7s",
		pb.description, pb.fill(), humanView(pb.cur), humanView(pb.contentLength))
	atomic.SwapInt32(&pb.printLenLine, int32(l))

	return
}

func humanView(num uint64) string {
	if num == 0 {
		return "0B"
	}

	baseUnit := 1000.0
	numFloat := float64(num)
	for _, unit := range []string{"B", "KB", "MB", "GB", "TB"} {
		if math.Abs(numFloat) < baseUnit {
			return fmt.Sprintf("%3.1f%s", numFloat, unit)
		}
		numFloat /= baseUnit
	}

	return fmt.Sprintf("%3.1fPiB", numFloat)
}

func NewProgressBar(width int8) *ProgressBar {
	return &ProgressBar{
		width: width,
	}
}
