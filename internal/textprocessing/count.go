package textprocessing

type counts struct {
	words   map[int64]int64
	ngram2  map[[2]int64]int64
	ngram3  map[[3]int64]int64
	ngram4  map[[4]int64]int64
	ngram5  map[[5]int64]int64
	ngram6  map[[6]int64]int64
	ngram7  map[[7]int64]int64
	ngram8  map[[8]int64]int64
	ngram9  map[[9]int64]int64
	ngram10 map[[10]int64]int64
}

func getCounts(tokens []int64, progress reportProgress) counts {

	counts := counts{
		words:   make(map[int64]int64),
		ngram2:  make(map[[2]int64]int64),
		ngram3:  make(map[[3]int64]int64),
		ngram4:  make(map[[4]int64]int64),
		ngram5:  make(map[[5]int64]int64),
		ngram6:  make(map[[6]int64]int64),
		ngram7:  make(map[[7]int64]int64),
		ngram8:  make(map[[8]int64]int64),
		ngram9:  make(map[[9]int64]int64),
		ngram10: make(map[[10]int64]int64),
	}

	countTokens := len(tokens)
	step := 1

	for end := 1; end <= countTokens; end++ {
		counts.words[tokens[end-1]]++

		if end >= 2 {
			counts.ngram2[[2]int64(tokens[end-2:end])]++
		}
		if end >= 3 {
			counts.ngram3[[3]int64(tokens[end-3:end])]++
		}
		if end >= 4 {
			counts.ngram4[[4]int64(tokens[end-4:end])]++
		}
		if end >= 5 {
			counts.ngram5[[5]int64(tokens[end-5:end])]++
		}
		if end >= 6 {
			counts.ngram6[[6]int64(tokens[end-6:end])]++
		}
		if end >= 7 {
			counts.ngram7[[7]int64(tokens[end-7:end])]++
		}
		if end >= 8 {
			counts.ngram8[[8]int64(tokens[end-8:end])]++
		}
		if end >= 9 {
			counts.ngram9[[9]int64(tokens[end-9:end])]++
		}
		if end >= 10 {
			counts.ngram10[[10]int64(tokens[end-10:end])]++
		}
		if step <= 10 && end == (countTokens*step)/10 {
			progress(step * 10)
			step++
		}
	}

	return counts

}
