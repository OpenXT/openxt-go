// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright 2026 Apertus Soutions, LLC
//

package ioctl

const (
	typBits = 8
	numBits = 8
	sizBits = 14
	dirBits = 2

	typMask = (1 << typBits) - 1
	numMask = (1 << numBits) - 1
	sizMask = (1 << sizBits) - 1
	dirMask = (1 << dirBits) - 1

	dirNone  = 0
	dirWrite = 1
	dirRead  = 2

	numShift = 0
	typShift = numShift + numBits
	sizShift = typShift + typBits
	dirShift = sizShift + sizBits
)

func Ioc(dir, t, nr, size uintptr) uintptr {
	return (dir << dirShift) | (t << typShift) | (nr << numShift) | (size << sizShift)
}

func Ior(t, nr, size uintptr) uintptr {
	return Ioc(dirRead, t, nr, size)
}

func Iow(t, nr, size uintptr) uintptr {
	return Ioc(dirWrite, t, nr, size)
}

func Iowr(t, nr, size uintptr) uintptr {
	return Ioc(dirRead|dirWrite, t, nr, size)
}
