package boltutil

import (
	"encoding/binary"
	"sort"
)

type intArr []int64

func (ia *intArr) GobDecode(b []byte) error {
	if len(b) < 8 {
		return nil
	}
	arr := *ia
	for i := 0; i < len(b); i += 8 {
		ogrn := binary.BigEndian.Uint64(b[i : i+8])
		arr = append(arr, int64(ogrn))
		//ch <- color.GreenString("%d %v", ogrn, ia)
	}
	ia = &arr
	return nil
}

func DecodeIntArr(b []byte) (intArr, error) {
	if len(b) < 8 {
		return nil, nil
	}
	arr := intArr{}
	for i := 0; i < len(b); i += 8 {
		ogrn := binary.BigEndian.Uint64(b[i : i+8])
		arr = append(arr, int64(ogrn))
		//ch <- color.GreenString("%d %v", ogrn, arr)
	}
	return arr, nil
}

func (ia intArr) GobEncode() ([]byte, error) {
	var out []byte

	//sort.IntSlice(ia)
	sort.Sort(ia)
	for _, i := range ia {
		bid := i2b(i)
		out = append(out, bid...)
	}
	return out, nil
}

func (ia intArr) Len() int           { return len(ia) }
func (ia intArr) Less(i, j int) bool { return ia[i] < ia[j] }
func (ia intArr) Swap(i, j int)      { ia[i], ia[j] = ia[j], ia[i] }

func (ia intArr) Has(i int64) bool {
	for _, v := range ia {
		if i == v {
			return true
		}
	}
	return false
}

func i2b(i int64) []byte {
	var bid = make([]byte, 8)
	binary.BigEndian.PutUint64(bid, uint64(i))
	return bid
}

func b2i(b []byte) int64 {
	u := binary.BigEndian.Uint64(b)
	return int64(u)
}
