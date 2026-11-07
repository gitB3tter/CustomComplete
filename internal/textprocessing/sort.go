package textprocessing

import (
	"slices"
)

type ngrams struct {
	words   [][2]int64
	ngram2  [][3]int64
	ngram3  [][4]int64
	ngram4  [][5]int64
	ngram5  [][6]int64
	ngram6  [][7]int64
	ngram7  [][8]int64
	ngram8  [][9]int64
	ngram9  [][10]int64
	ngram10 [][11]int64
}

func getSorted(c counts, progress reportProgress) ngrams {

	u := getUnsorted(c)

	slices.SortFunc(u.words, func(a, b [2]int64) int {
		return slices.Compare(a[:], b[:])
	})
	progress(10)
	slices.SortFunc(u.ngram2, func(a, b [3]int64) int {
		return slices.Compare(a[:], b[:])
	})
	progress(20)
	slices.SortFunc(u.ngram3, func(a, b [4]int64) int {
		return slices.Compare(a[:], b[:])
	})
	progress(30)
	slices.SortFunc(u.ngram4, func(a, b [5]int64) int {
		return slices.Compare(a[:], b[:])
	})
	progress(40)
	slices.SortFunc(u.ngram5, func(a, b [6]int64) int {
		return slices.Compare(a[:], b[:])
	})
	progress(50)
	slices.SortFunc(u.ngram6, func(a, b [7]int64) int {
		return slices.Compare(a[:], b[:])
	})
	progress(60)
	slices.SortFunc(u.ngram7, func(a, b [8]int64) int {
		return slices.Compare(a[:], b[:])
	})
	progress(70)
	slices.SortFunc(u.ngram8, func(a, b [9]int64) int {
		return slices.Compare(a[:], b[:])
	})
	progress(80)
	slices.SortFunc(u.ngram9, func(a, b [10]int64) int {
		return slices.Compare(a[:], b[:])
	})
	progress(90)
	slices.SortFunc(u.ngram10, func(a, b [11]int64) int {
		return slices.Compare(a[:], b[:])
	})

	return u
}

func getUnsorted(c counts) ngrams {

	u := ngrams{}
	for word, count := range c.words {
		u.words = append(u.words, [2]int64{word, count})
	}
	for ngram2, count := range c.ngram2 {
		u.ngram2 = append(u.ngram2, [3]int64{ngram2[0], ngram2[1], count})
	}
	for ngram3, count := range c.ngram3 {
		u.ngram3 = append(u.ngram3, [4]int64{ngram3[0], ngram3[1], ngram3[2], count})
	}
	for ngram4, count := range c.ngram4 {
		u.ngram4 = append(u.ngram4, [5]int64{ngram4[0], ngram4[1], ngram4[2], ngram4[3], count})
	}
	for ngram5, count := range c.ngram5 {
		u.ngram5 = append(u.ngram5, [6]int64{ngram5[0], ngram5[1], ngram5[2], ngram5[3], ngram5[4], count})
	}
	for ngram6, count := range c.ngram6 {
		u.ngram6 = append(u.ngram6, [7]int64{ngram6[0], ngram6[1], ngram6[2], ngram6[3], ngram6[4], ngram6[5], count})
	}
	for ngram7, count := range c.ngram7 {
		u.ngram7 = append(u.ngram7, [8]int64{ngram7[0], ngram7[1], ngram7[2], ngram7[3], ngram7[4], ngram7[5], ngram7[6], count})
	}
	for ngram8, count := range c.ngram8 {
		u.ngram8 = append(u.ngram8, [9]int64{ngram8[0], ngram8[1], ngram8[2], ngram8[3], ngram8[4], ngram8[5], ngram8[6], ngram8[7], count})
	}
	for ngram9, count := range c.ngram9 {
		u.ngram9 = append(u.ngram9, [10]int64{ngram9[0], ngram9[1], ngram9[2], ngram9[3], ngram9[4], ngram9[5], ngram9[6], ngram9[7], ngram9[8], count})
	}
	for ngram10, count := range c.ngram10 {
		u.ngram10 = append(u.ngram10, [11]int64{ngram10[0], ngram10[1], ngram10[2], ngram10[3], ngram10[4], ngram10[5], ngram10[6], ngram10[7], ngram10[8], ngram10[9], count})
	}

	return u
}
