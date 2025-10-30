package main

import (
	mapset "github.com/deckarep/golang-set/v2"
)

func collectValuesAsSet[U comparable, T any](slice []T, f func(T) U) mapset.Set[U] {
	s := mapset.NewSet[U]()
	for _, e := range slice {
		s.Add(f(e))
	}
	return s
}
