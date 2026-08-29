// Copyright (C) 1993 by Sun Microsystems, Inc. All rights reserved.
//
// Developed at SunSoft, a Sun Microsystems, Inc. business.
// Permission to use, copy, modify, and distribute this software is freely
// granted, provided that this notice is preserved.

package ecmascript

// cspell:ignore fdlibm

import "math"

// Acos implements the fdlibm algorithm used by V8's Math.acos. Go's
// math.Acos differs from V8 by one ULP for some inputs, which is observable to
// JavaScript constant evaluation through strict equality.
func Acos(x float64) float64 {
	const (
		one    = 1.00000000000000000000e+00
		pi     = 3.14159265358979311600e+00
		pio2Hi = 1.57079632679489655800e+00
		pio2Lo = 6.12323399573676603587e-17
		pS0    = 1.66666666666666657415e-01
		pS1    = -3.25565818622400915405e-01
		pS2    = 2.01212532134862925881e-01
		pS3    = -4.00555345006794114027e-02
		pS4    = 7.91534994289814532176e-04
		pS5    = 3.47933107596021167570e-05
		qS1    = -2.40339491173441421878e+00
		qS2    = 2.02094576023350569471e+00
		qS3    = -6.88283971605453293030e-01
		qS4    = 7.70381505559019352791e-02
	)

	bits := math.Float64bits(x)
	high := int32(bits >> 32)
	absHigh := high & 0x7fffffff
	if absHigh >= 0x3ff00000 {
		low := uint32(bits)
		if (uint32(absHigh-0x3ff00000) | low) == 0 {
			if high > 0 {
				return 0
			}
			return pi + 2*pio2Lo
		}
		return math.NaN()
	}
	if absHigh < 0x3fe00000 {
		if absHigh <= 0x3c600000 {
			return pio2Hi + pio2Lo
		}
		z := x * x
		p := z * (pS0 + z*(pS1+z*(pS2+z*(pS3+z*(pS4+z*pS5)))))
		q := one + z*(qS1+z*(qS2+z*(qS3+z*qS4)))
		r := p / q
		return pio2Hi - (x - (pio2Lo - x*r))
	}
	if high < 0 {
		z := (one + x) * 0.5
		p := z * (pS0 + z*(pS1+z*(pS2+z*(pS3+z*(pS4+z*pS5)))))
		q := one + z*(qS1+z*(qS2+z*(qS3+z*qS4)))
		s := math.Sqrt(z)
		r := p / q
		w := r*s - pio2Lo
		return pi - 2*(s+w)
	}
	z := (one - x) * 0.5
	s := math.Sqrt(z)
	df := math.Float64frombits(math.Float64bits(s) & 0xffffffff00000000)
	c := (z - df*df) / (s + df)
	p := z * (pS0 + z*(pS1+z*(pS2+z*(pS3+z*(pS4+z*pS5)))))
	q := one + z*(qS1+z*(qS2+z*(qS3+z*qS4)))
	r := p / q
	w := r*s + c
	return 2 * (df + w)
}
