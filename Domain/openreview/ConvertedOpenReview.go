package openreview

import (
	"fmt"
	"log"
	"regexp"
	"sort"
)

type ConvertedOpenReview struct{}

func NewConvertedOpenReview() *ConvertedOpenReview {
	return &ConvertedOpenReview{}
}

func (c *ConvertedOpenReview) Convert(sources []string) ([]string, error) {
	if len(sources) == 0 {
		return nil, fmt.Errorf("変換対象がありません")
	}

	idx := sort.Search(len(sources), func(i int) bool {
		return sources[i] >= "%"
	})
	fmt.Print(idx)

	var canSkipLine bool
	canAppendPrefix := true
	canSearchedMachiningONumber := true
	regFdNo := regexp.MustCompile(`^O100[12]$`)
	rexM00 := regexp.MustCompile(`^M00$`)
	rexM01 := regexp.MustCompile(`^M01$`)
	rexM30orM99 := regexp.MustCompile(`^\(M(30|99)\)$`)
	rexM30 := regexp.MustCompile(`^M30$`)
	canAppendFinallyM30 := true
	regPercentOrBlank := regexp.MustCompile(`^%?$`)
	regPercent := regexp.MustCompile(`^%$`)
	regBlank := regexp.MustCompile(`^$`)
	var res []string
	for i, line := range sources {
		log.Println("line:", line, "canSkipLine:", canSkipLine)

		// オープンレビューのファイルの先頭つける予約語
		if canAppendPrefix && !regBlank.MatchString(line) {
			res = append([]string{"%", "O1002"}, res...)
			canAppendPrefix = false
			if regPercent.MatchString(line) {
				// %の場合
				continue
			}
		}

		if canSearchedMachiningONumber && regFdNo.MatchString(line) {
			// 他の命令が出てくる前に、Oナンバーがあったら、無視(消す)
			canSearchedMachiningONumber = false
			continue
		}

		if rexM00.MatchString(line) {
			canSearchedMachiningONumber = false
			continue
		}

		if !canSkipLine && rexM01.MatchString(line) {
			res = append(res, line)
			canSkipLine = true
			canSearchedMachiningONumber = false
		} else if canSkipLine && rexM30orM99.MatchString(line) {
			res = append(res, line)
			canSkipLine = false
			canSearchedMachiningONumber = false
		} else if !canSkipLine {
			if i > 0 && canAppendFinallyM30 && rexM30.MatchString(line) {
				// M30が見つかったら追記しなくても良いかも
				canAppendFinallyM30 = false
			} else if i > 0 && !canAppendFinallyM30 && !regPercentOrBlank.MatchString(line) {
				// '?' '空行' 以外が見つかったら最後のコマンドじゃなかった
				canAppendFinallyM30 = true
			}

			if len(line) > 0 {
				// 行頭から命令があったら もう検索しない
				canSearchedMachiningONumber = false
			}

			if i > 0 && canAppendFinallyM30 && regPercent.MatchString(line) {
				// M30はファイルの最後の%の1つ前に追記する
				res = append(res, "M30")
			}
			res = append(res, line)
		}
	}
	return res, nil
}
