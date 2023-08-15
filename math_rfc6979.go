package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"hash"
	"math/big"
)

// rfc6979 implemented in Golang.
// copy from  https://raw.githubusercontent.com/codahale/rfc6979/master/rfc6979.go
/*
Package rfc6979 is an implementation of RFC 6979's deterministic DSA.
	Such signatures are compatible with standard Digital Signature Algorithm
	(DSA) and Elliptic Curve Digital Signature Algorithm (ECDSA) digital
	signatures and can be processed with unmodified verifiers, which need not be
	aware of the procedure described therein.  Deterministic signatures retain
	the cryptographic security features associated with digital signatures but
	can be more easily implemented in various environments, since they do not
	need access to a source of high-quality randomness.
(https://tools.ietf.org/html/rfc6979)
Provides functions similar to crypto/dsa and crypto/ecdsa.
*/

// mac returns an HMAC of the given key and message.
func mac(alg func() hash.Hash, k, m, buf []byte) []byte {
	h := hmac.New(alg, k)
	h.Write(m)
	return h.Sum(buf[:0])
}

// https://tools.ietf.org/html/rfc6979#section-2.3.2
func bits2int(in []byte, qlen int) *big.Int {
	vlen := len(in) * 8
	v := new(big.Int).SetBytes(in)
	if vlen > qlen {
		v = new(big.Int).Rsh(v, uint(vlen-qlen))
	}
	return v
}

// https://tools.ietf.org/html/rfc6979#section-2.3.3
func int2octets(v *big.Int, rolen int) []byte {
	out := v.Bytes()

	// pad with zeros if it's too short
	if len(out) < rolen {
		out2 := make([]byte, rolen)
		copy(out2[rolen-len(out):], out)
		return out2
	}

	// drop most significant bytes if it's too long
	if len(out) > rolen {
		out2 := make([]byte, rolen)
		copy(out2, out[len(out)-rolen:])
		return out2
	}

	return out
}

// https://tools.ietf.org/html/rfc6979#section-2.3.4
func bits2octets(in []byte, q *big.Int, qlen, rolen int) []byte {
	z1 := bits2int(in, qlen)
	z2 := new(big.Int).Sub(z1, q)
	if z2.Sign() < 0 {
		return int2octets(z1, rolen)
	}
	return int2octets(z2, rolen)
}

// https://tools.ietf.org/html/rfc6979#section-3.2
func generateSecret(q, x *big.Int, alg func() hash.Hash, hash []byte, extraEntropy []byte) *big.Int {
	qlen := q.BitLen()
	holen := alg().Size()
	rolen := (qlen + 7) >> 3
	bx := append(int2octets(x, rolen), bits2octets(hash, q, qlen, rolen)...)
	// extra_entropy - extra added data in binary form as per section-3.6 of rfc6979
	if len(extraEntropy) > 0 {
		bx = append(bx, extraEntropy...)
	}

	// Step B
	v := bytes.Repeat([]byte{0x01}, holen)

	// Step C
	k := bytes.Repeat([]byte{0x00}, holen)

	// Step D
	k = mac(alg, k, append(append(v, 0x00), bx...), k)

	// Step E
	v = mac(alg, k, v, v)

	// Step F
	k = mac(alg, k, append(append(v, 0x01), bx...), k)

	// Step G
	v = mac(alg, k, v, v)

	// Step H
	for {
		// Step H1
		var t []byte

		// Step H2
		for len(t) < qlen/8 {
			v = mac(alg, k, v, v)
			t = append(t, v...)
		}

		// Step H3
		secret := bits2int(t, qlen)
		if secret.Cmp(one) >= 0 && secret.Cmp(q) < 0 {
			return secret
		}
		k = mac(alg, k, append(v, 0x00), k)
		v = mac(alg, k, v, v)
	}
}

func GenerateKRfc6979(msgHash, priKey *big.Int, seed int) *big.Int {
	msgHash = big.NewInt(0).Set(msgHash) // copy
	bitMod := msgHash.BitLen() % 8
	if bitMod <= 4 && bitMod >= 1 && msgHash.BitLen() > 248 {
		msgHash.Mul(msgHash, big.NewInt(16))
	}
	var extra []byte
	if seed > 0 {
		buf := new(bytes.Buffer)
		var data interface{}
		if seed < 256 {
			data = uint8(seed)
		} else if seed < 65536 {
			data = uint16(seed)
		} else if seed < 4294967296 {
			data = uint32(seed)
		} else {
			data = uint64(seed)
		}
		_ = binary.Write(buf, binary.BigEndian, data)
		extra = buf.Bytes()
	}
	return generateSecret(EC_ORDER, priKey, sha256.New, msgHash.Bytes(), extra)
}
func doSign(s1 string, s2 string) (*big.Int, *big.Int) {
	priKey, _ := new(big.Int).SetString(s1, 16)
	msgHash, _ := new(big.Int).SetString(s2, 10)
	seed := 0
	EcGen := pedersenCfg.ConstantPoints[1]
	alpha := pedersenCfg.ALPHA
	nBit := big.NewInt(0).Exp(big.NewInt(2), N_ELEMENT_BITS_ECDSA, nil)
	for {
		k := GenerateKRfc6979(msgHash, priKey, seed)
		//	Update seed for next iteration in case the value of k is bad.
		if seed == 0 {
			seed = 1
		} else {
			seed += 1
		}
		// Cannot fail because 0 < k < EC_ORDER and EC_ORDER is prime.
		x := ecMult(k, EcGen, alpha, FIELD_PRIME)[0]
		// !(1 <= x < 2 ** N_ELEMENT_BITS_ECDSA)
		if !(x.Cmp(one) > 0 && x.Cmp(nBit) < 0) {
			continue
		}
		// msg_hash + r * priv_key
		x1 := big.NewInt(0).Add(msgHash, big.NewInt(0).Mul(x, priKey))
		// (msg_hash + r * priv_key) % EC_ORDER == 0
		if big.NewInt(0).Mod(x1, EC_ORDER).Cmp(zero) == 0 {
			continue
		}
		// w = div_mod(k, msg_hash + r * priv_key, EC_ORDER)
		w := divMod(k, x1, EC_ORDER)
		// not (1 <= w < 2 ** N_ELEMENT_BITS_ECDSA)
		if !(w.Cmp(one) > 0 && w.Cmp(nBit) < 0) {
			continue
		}
		s1 := divMod(one, w, EC_ORDER)
		return x, s1
	}
}
