// Package set provides both threadsafe and non-threadsafe implementations of
// a generic set data structure. In the threadsafe set, safety encompasses all
// operations on one set. Operations on multiple sets are consistent in that
// the elements of each set used was valid at exactly one point in time
// between the start and the end of the operation.
package set

// Interface is describing a Set. Sets are an unordered, unique list of values.
type Interface interface {
	New(items ...string) Interface
	Add(items ...string)
	Remove(items ...string)
	Pop() string
	Has(items ...string) bool
	Size() int
	Clear()
	IsEmpty() bool
	IsEqual(s Interface) bool
	IsSubset(s Interface) bool
	IsSuperset(s Interface) bool
	Each(func(string) bool)
	String() string
	List() []string
	Copy() Interface
	Merge(s Interface)
	Separate(s Interface)
}

// helpful to not write everywhere struct{}{}
var keyExists = struct{}{}

// Union is the merger of multiple sets. It returns a new set with all the
// elements present in all the sets that are passed.
//
// The dynamic type of the returned set is determined by the first passed set's
// implementation of the New() method.
func Union(set1, set2 Interface, sets ...Interface) Interface {
	u := set1.Copy()
	set2.Each(func(item string) bool {
		u.Add(item)
		return true
	})
	for _, set := range sets {
		set.Each(func(item string) bool {
			u.Add(item)
			return true
		})
	}

	return u
}

// Difference returns a new set which contains items which are in in the first
// set but not in the others. Unlike the Difference() method you can use this
// function separately with multiple sets.
func Difference(set1, set2 Interface, sets ...Interface) Interface {
	s := set1.Copy()
	s.Separate(set2)
	for _, set := range sets {
		s.Separate(set) // seperate is thread safe
	}
	return s
}

// Intersection returns a new set which contains items that only exist in all given sets.
func Intersection(set1, set2 Interface, sets ...Interface) Interface {
	all := Union(set1, set2, sets...)
	result := Union(set1, set2, sets...)

	all.Each(func(item string) bool {
		if !set1.Has(item) || !set2.Has(item) {
			result.Remove(item)
		}

		for _, set := range sets {
			if !set.Has(item) {
				result.Remove(item)
			}
		}
		return true
	})
	return result
}

// SymmetricDifference returns a new set which s is the difference of items which are in
// one of either, but not in both.
func SymmetricDifference(s Interface, t Interface) Interface {
	u := Difference(s, t)
	v := Difference(t, s)
	return Union(u, v)
}
