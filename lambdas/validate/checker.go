package main

import (
	"strconv"
	"time"
)

type SingleColumnChecker struct {
	idx   int
	check func(*string) bool
}

var singleColumnCheckers = []SingleColumnChecker{
	{
		idx: 0,
		check: func(s *string) bool {
			i, err := strconv.ParseInt(*s, 10, 32)
			if err != nil || i < 0 {
				*s = ""
			}

			return true
		},
	},
	{
		idx: 1,
		check: func(s *string) bool {
			i, err := strconv.ParseInt(*s, 10, 32)
			if err != nil || i < 1 || i > 2 {
				*s = ""
			}

			return true
		},
	},
	{
		idx: 4,
		check: func(s *string) bool {
			i, err := strconv.ParseFloat(*s, 32)
			if err != nil {
				*s = ""
			} else if i < 1 || i > 5 {
				return false
			}

			return true
		},
	},
	{
		idx: 5,
		check: func(s *string) bool {
			i, err := strconv.ParseFloat(*s, 32)
			if err != nil || i < 0 {
				return false
			}

			return true
		},
	},
	{
		idx: 6,
		check: func(s *string) bool {
			i, err := strconv.ParseFloat(*s, 32)
			if err != nil || i < 0 || (i > 6 && i != 99) {
				return false
			}

			return true
		},
	},
	{
		idx: 7,
		check: func(s *string) bool {
			if *s != "Y" && *s != "N" {
				*s = ""
			}

			return true
		},
	},
	{
		idx: 8,
		check: func(s *string) bool {
			i, err := strconv.ParseInt(*s, 10, 32)
			if err != nil || i < 0 {
				return false
			}

			return true
		},
	},
	{
		idx: 9,
		check: func(s *string) bool {
			i, err := strconv.ParseInt(*s, 10, 32)
			if err != nil || i < 0 {
				return false
			}

			return true
		},
	},
	{
		idx: 10,
		check: func(s *string) bool {
			i, err := strconv.ParseInt(*s, 10, 32)
			if err != nil || i < 0 || i > 6 {
				*s = ""
			}

			return true
		},
	},
	{
		idx: 11,
		check: func(s *string) bool {
			_, err := strconv.ParseFloat(*s, 32)
			if err != nil {
				*s = ""
			}

			return true
		},
	},
	{
		idx: 12,
		check: func(s *string) bool {
			_, err := strconv.ParseFloat(*s, 32)
			if err != nil {
				*s = ""
			}

			return true
		},
	},
	{
		idx: 13,
		check: func(s *string) bool {
			_, err := strconv.ParseFloat(*s, 32)
			if err != nil {
				*s = ""
			}

			return true
		},
	},
	{
		idx: 14,
		check: func(s *string) bool {
			_, err := strconv.ParseFloat(*s, 32)
			if err != nil {
				*s = ""
			}

			return true
		},
	},
	{
		idx: 15,
		check: func(s *string) bool {
			_, err := strconv.ParseFloat(*s, 32)
			if err != nil {
				*s = ""
			}

			return true
		},
	},
	{
		idx: 16,
		check: func(s *string) bool {
			_, err := strconv.ParseFloat(*s, 32)
			if err != nil {
				*s = ""
			}

			return true
		},
	},
	{
		idx: 17,
		check: func(s *string) bool {
			i, err := strconv.ParseFloat(*s, 32)
			if err != nil || i == 0 {
				return false
			}

			return true
		},
	},
	{
		idx: 18,
		check: func(s *string) bool {
			_, err := strconv.ParseFloat(*s, 32)
			if err != nil {
				*s = ""
			}

			return true
		},
	},
	{
		idx: 19,
		check: func(s *string) bool {
			_, err := strconv.ParseFloat(*s, 32)
			if err != nil {
				*s = ""
			}

			return true
		},
	},
}

type CrossColumnChecker struct {
	idxs  []int
	check func(*[]*string) bool
}

var crossColumnCheckers = []CrossColumnChecker{
	{
		idxs: []int{2, 3},
		check: func(cols *[]*string) bool {
			layoutDate := "2019-05-13 23:59:59"

			d1, d1err := time.Parse(layoutDate, *(*cols)[0])
			d2, d2err := time.Parse(layoutDate, *(*cols)[1])

			if d1err != nil || d2err != nil {
				return false
			}

			return d1.Compare(d2) == -1
		},
	},
	{
		idxs: []int{10, 11, 12, 13, 14, 15, 16, 17, 18, 19},
		check: func(cols *[]*string) bool {
			if *(*cols)[0] == "3" {
				for _, e := range (*cols)[1:] {
					if len(*e) == 0 {
						return true
					}
				}

				return false
			}

			return true
		},
	},
	{
		idxs: []int{11, 12, 13, 14, 15, 16, 17, 18, 19},
		check: func(cols *[]*string) bool {

		},
	},
}
