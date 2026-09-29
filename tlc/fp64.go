package tlc

import (
	"io"
	"math/rand"
	"time"
	"unicode/utf16"
)

const (
	fp64One  = uint64(0x8000000000000000)
	fp64X63  = uint64(0x1)
	FP64Zero = uint64(0)
)

var FP64Polys = []uint64{
	0x911498AE0E66BAD6,
	0xda8a0ba66dae0181,
	0xc02f176b8f268d9f,
	0xd617bb1220fc7812,
	0xc6fd951ad34f9f74,
	0xdd1897bd991704d4,
	0xf5394c541cbfd343,
	0xb1dded37b5c7b8f7,
	0xb713ff61039dc632,
	0xdfb340cb2fb03d43,
	0xbc3e7e4c5ecb76a3,
	0xdbb4b1349cd7058a,
	0xf53e9dcb9e915cdf,
	0xca5f58e90dd01848,
	0x80e7ff4406891aa1,
	0xab541bf881fa8571,
	0xbf274e07ac5499d5,
	0x939b1ea933040a4e,
	0xb791a595448d75b1,
	0x8bf88d6ef85563a2,
	0xecb33ec339513a53,
	0xfa2d3e722db5208a,
	0xb4e2058aac479d24,
	0xafbd6474e7213b82,
	0x98c1694d14ffaeef,
	0xe188fb5c0a125e24,
	0xfa71cc3865487d80,
	0x891135f7c1c94569,
	0xcf77cdd16d22e3e6,
	0xeb5e3a1d2e2bb4b5,
	0x92f7f5b69cd00c55,
	0xa9fbbe40ca3b9ae9,
	0x84a7b33d85295bde,
	0xebd4680dbb6fdee2,
	0xa31fc46a0583b4d0,
	0xa792c94f15de3e49,
	0xd9d60a9feff4521a,
	0x9227ba31dfdda04b,
	0xfb4c89c607ce162d,
	0xa89b3b2e01479cc0,
	0xb35a0c2a28b89f7d,
	0x91d0b700b99d9ec2,
	0xf0646bbda05020b0,
	0xcb5d5f63ce043056,
	0xd276b6a04f42a1b8,
	0xbc1a7a7dbfeb47a3,
	0xe138acff7a963036,
	0xed860223c1557ee7,
	0x9b2491e980150ab6,
	0xe7e03dd8a5b4e59e,
	0xaaa3f5eac516783c,
	0xfe78cc267a724180,
	0xc22519a21edfac64,
	0xcdb2941933fec60f,
	0xc5f485551ef38aa1,
	0xa19293d250bd3335,
	0xa4d4c215a50b7afe,
	0xc1155176406a5070,
	0xecadaa8200e123ec,
	0xbacfe629d58b2f08,
	0xded991082148cb42,
	0x9a0ccbe5deaa88db,
	0x9a83246f342061bb,
	0xb71482842297ab05,
	0xf2407eeb997592bd,
	0xb7b43d4b5c4bcc,
	0xb339a2568221ffe7,
	0xdb4b6b379446ef9e,
	0xe43c205bb5c0b2b6,
	0xe8e1d141f19d6db0,
	0xd19e8710a4ea1c86,
	0x9704cecfa8d6d07b,
	0xb0a35716162c3f26,
	0xa2a68c0cf56390ac,
	0xe4a74bc601c95b46,
	0xe668fff675595e56,
	0xe0ea77ebe06fabbb,
	0xa8bb94f585279523,
	0xbeb667b42b684f21,
	0xd7b65410189e28af,
	0x85722037beffc5b9,
	0xe7e7c5f773426204,
	0xa5fd0cc8e060c6f4,
	0xb8f91ee9065dcf95,
	0xb4047008040f8b50,
	0xecb9ab6291c8cfcc,
	0xe08bb9b70caad6df,
	0xd6e086a301d95d56,
	0xf6a808f5f3fb9da3,
	0xfa74b8a8ef86fbc5,
	0xa0b6b33ba9e6381c,
	0x8c78703427873dbb,
	0xc5516ea423011021,
	0xb075cce8528ae7e2,
	0x92e4a37979e2b13b,
	0xafc9ab000ed81026,
	0xf1873f2a861a518a,
	0xf885d6b35770192c,
	0xd2a82f27f71f5f7e,
	0xd5e7a4dd8beb2d9f,
	0xbab9e7e65dc23e0f,
	0xd5fc877c0bdf5b85,
	0xab428169b9f31c02,
	0xb7b1351f9266d3ea,
	0xfad564914328f635,
	0x9623ecb1000db9bd,
	0x88d371ec6644c892,
	0xd0a71270573e271c,
	0xbd200d763f9d81b4,
	0x8484fd96374bf2ea,
	0xe0d0810749432294,
	0x8ebe9dadc88658ba,
	0xf6268c58993ae542,
	0xd6cc88fed0d359c9,
	0xbcaddb8d40a16690,
	0x92817c6db421cbf6,
	0xac63721120a371b5,
	0xd6cc137fdced0820,
	0xf1c5fbaebb617bd7,
	0xac35d78c765237a5,
	0xbd0a471fa9a23116,
	0x943a7031b946a5ae,
	0xe4e83520cb1aaebb,
	0xe92a4b73246dd124,
	0xafc1e070787c4c86,
	0xdfe84d42cf06286b,
	0x8b29ec962e4b964b,
	0x807eb5de812ede0f,
	0xa3cd71299c8b3bfd,
	0x845b8031ef886f35,
	0x91f5a5fa9c5515a5,
}

var (
	fp64ByteModTable7 [256]uint64
	fp64IrredPoly     uint64
)

func FP64IrredPoly() uint64 {
	return fp64IrredPoly
}

func FP64InitRandom() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	FP64InitIndex(r.Intn(len(FP64Polys)))
}

func FP64Init() {
	FP64InitIndex(0)
}

func FP64InitIndex(n int) {
	FP64InitPoly(FP64Polys[n])
}

func FP64InitPoly(poly uint64) {
	fp64IrredPoly = poly
	powerTable := make([]uint64, 72)

	t := fp64One
	for i := range powerTable {
		powerTable[i] = t
		mask := uint64(0)
		if t&fp64X63 != 0 {
			mask = fp64IrredPoly
		}
		t = (t >> 1) ^ mask
	}

	for j := 0; j <= 255; j++ {
		v := uint64(0)
		for k := 0; k <= 7; k++ {
			if j&(1<<k) != 0 {
				v ^= powerTable[127-(7*8)-k]
			}
		}
		fp64ByteModTable7[j] = v
	}
}

func FP64New() uint64 {
	return fp64IrredPoly
}

func FP64NewString(s string) uint64 {
	return FP64ExtendString(fp64IrredPoly, s)
}

func FP64NewBytes(bytes []byte) uint64 {
	return FP64ExtendBytes(fp64IrredPoly, bytes)
}

func FP64NewReader(r io.Reader) (uint64, error) {
	return FP64ExtendReader(fp64IrredPoly, r)
}

func FP64ExtendString(fp uint64, s string) uint64 {
	return FP64ExtendUTF16(fp, utf16.Encode([]rune(s)))
}

func FP64ExtendUTF16(fp uint64, chars []uint16) uint64 {
	for _, c := range chars {
		fp = (fp >> 8) ^ fp64ByteModTable7[(uint16(fp)^c)&0xff]
	}
	return fp
}

func FP64ExtendBytes(fp uint64, bytes []byte) uint64 {
	for _, b := range bytes {
		fp = (fp >> 8) ^ fp64ByteModTable7[(uint64(b)^fp)&0xff]
	}
	return fp
}

func FP64ExtendReader(fp uint64, r io.Reader) (uint64, error) {
	var buf [8192]byte
	for {
		n, err := r.Read(buf[:])
		if n > 0 {
			fp = FP64ExtendBytes(fp, buf[:n])
		}
		if err == io.EOF {
			return fp, nil
		}
		if err != nil {
			return fp, err
		}
	}
}

func FP64ExtendByte(fp uint64, b byte) uint64 {
	return (fp >> 8) ^ fp64ByteModTable7[(uint64(b)^fp)&0xff]
}

func FP64ExtendInt(fp uint64, x int32) uint64 {
	u := uint32(x)
	for i := 0; i < 4; i++ {
		fp = FP64ExtendByte(fp, byte(u&0xff))
		u >>= 8
	}
	return fp
}

func FP64ExtendLong(fp uint64, x int64) uint64 {
	u := uint64(x)
	for i := 0; i < 8; i++ {
		fp = FP64ExtendByte(fp, byte(u&0xff))
		u >>= 8
	}
	return fp
}

func FP64Hash(fp uint64) int32 {
	return int32(fp)
}
