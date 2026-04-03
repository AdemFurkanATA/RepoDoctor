package main

import (
	"reflect"
	"sort"
	"testing"
	"testing/quick"

	"RepoDoctor/internal/model"
)

func TestProperty_SortedStringCopy_IdempotentAndSorted(t *testing.T) {
	cfg := &quick.Config{MaxCount: 200}
	prop := func(input []string) bool {
		original := append([]string(nil), input...)
		got := sortedStringCopy(input)

		if !sort.StringsAreSorted(got) {
			return false
		}

		if !equalStringSlices(input, original) {
			return false
		}

		got2 := sortedStringCopy(got)
		return equalStringSlices(got, got2)
	}

	if err := quick.Check(prop, cfg); err != nil {
		t.Fatalf("sortedStringCopy property failed: %v", err)
	}
}

func equalStringSlices(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func TestProperty_StableSortedCopy_Ints_MatchesGoSortAndIdempotent(t *testing.T) {
	cfg := &quick.Config{MaxCount: 200}
	prop := func(input []int) bool {
		got := stableSortedCopy(input, func(left, right int) bool { return left < right })
		want := append([]int(nil), input...)
		sort.Ints(want)

		if !reflect.DeepEqual(got, want) {
			return false
		}

		got2 := stableSortedCopy(got, func(left, right int) bool { return left < right })
		return reflect.DeepEqual(got, got2)
	}

	if err := quick.Check(prop, cfg); err != nil {
		t.Fatalf("stableSortedCopy property failed: %v", err)
	}
}

func TestProperty_SortModelViolations_DeterministicAcrossPermutations(t *testing.T) {
	cfg := &quick.Config{MaxCount: 150}
	prop := func(a, b, c, d string, l1, l2 uint8) bool {
		base := []model.Violation{
			{RuleID: a, File: b, Line: int(l1), Message: c},
			{RuleID: b, File: c, Line: int(l2), Message: d},
			{RuleID: a, File: c, Line: int(l2), Message: b},
			{RuleID: d, File: a, Line: int(l1), Message: a},
		}

		x := append([]model.Violation(nil), base...)
		y := []model.Violation{base[3], base[1], base[0], base[2]}

		sortModelViolationsDeterministic(x)
		sortModelViolationsDeterministic(y)

		if !reflect.DeepEqual(x, y) {
			return false
		}

		z := append([]model.Violation(nil), x...)
		sortModelViolationsDeterministic(z)
		return reflect.DeepEqual(x, z)
	}

	if err := quick.Check(prop, cfg); err != nil {
		t.Fatalf("sortModelViolationsDeterministic property failed: %v", err)
	}
}
