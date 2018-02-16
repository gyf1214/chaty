package util

import (
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"math/big"
)

var curve elliptic.Curve

func init() {
	curve = elliptic.P256()
}

type Point struct {
	x, y *big.Int
}

func PointFromDecode(str string) (Point, error) {
	decoded, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return Point{}, err
	}

	x, y := elliptic.Unmarshal(curve, decoded)
	if x == nil {
		return Point{}, errors.New("parse error")
	}

	return Point{x: x, y: y}, nil
}

func PointFromRandom() (Point, error) {
	_, x, y, err := elliptic.GenerateKey(curve, rand.Reader)
	return Point{x: x, y: y}, err
}

func PointFromPriv(priv []byte) (Point, error) {
	x, y := curve.ScalarBaseMult(priv)
	if x == nil {
		return Point{}, errors.New("priv error")
	}
	return Point{x: x, y: y}, nil
}

var mask = []byte{0xff, 0x1, 0x3, 0x7, 0xf, 0x1f, 0x3f, 0x7f}

func RandomFromCurve() ([]byte, error) {
	n := curve.Params().N
	bits := n.BitLen()
	bytes := (bits + 7) >> 3
	ret := make([]byte, bytes)

	for {
		_, err := io.ReadFull(rand.Reader, ret)
		if err != nil {
			return nil, err
		}

		ret[0] &= mask[bits%8]
		ret[1] ^= 0x42
		if new(big.Int).SetBytes(ret).Cmp(n) < 0 {
			return ret, nil
		}
	}
}

func (p Point) Encode() string {
	data := elliptic.Marshal(curve, p.x, p.y)
	return base64.StdEncoding.EncodeToString(data)
}

func (p Point) Hash() []byte {
	data := elliptic.Marshal(curve, p.x, p.y)
	ret := sha256.Sum256(data)
	return ret[:]
}

func (p Point) Add(o Point) Point {
	x, y := curve.Add(p.x, p.y, o.x, o.y)
	return Point{x: x, y: y}
}

func (p Point) Neg() Point {
	return Point{
		x: (new(big.Int)).Set(p.x),
		y: (new(big.Int)).Neg(p.y),
	}
}

func (p Point) Mul(o []byte) Point {
	x, y := curve.ScalarMult(p.x, p.y, o)
	return Point{x: x, y: y}
}
