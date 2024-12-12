// Copyright 2013-2015 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package semver

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type Version struct {
	Patched    []int64
	Original   string
	PreRelease PreRelease
	Metadata   string
}

type PreRelease string

func splitOff(input *string, delim string) (val string) {
	parts := strings.SplitN(*input, delim, 2)

	if len(parts) == 2 {
		*input = parts[0]
		val = parts[1]
	}

	return val
}

func New(version string) *Version {
	return Must(NewVersion(version))
}

func NewVersion(version string) (*Version, error) {
	v := Version{}

	if err := v.Set(version); err != nil {
		return nil, err
	}

	return &v, nil
}

// Must is a helper for wrapping NewVersion and will panic if err is not nil.
func Must(v *Version, err error) *Version {
	if err != nil {
		panic(err)
	}
	return v
}

// Set parses and updates v from the given version string. Implements flag.Value
func (v *Version) Set(version string) error {
	v.Original = version
	metadata := splitOff(&version, "+")
	preRelease := PreRelease(splitOff(&version, "-"))
	dotParts := strings.Split(version, ".")

	if err := validateIdentifier(string(preRelease)); err != nil {
		return fmt.Errorf("failed to validate pre-release: %v", err)
	}

	if err := validateIdentifier(metadata); err != nil {
		return fmt.Errorf("failed to validate metadata: %v", err)
	}

	v.Patched = make([]int64, len(dotParts))
	for i, dd := range dotParts {
		val, err := strconv.ParseInt(dd, 10, 64)
		v.Patched[i] = val
		if err != nil {
			return err
		}
	}

	v.Metadata = metadata
	v.PreRelease = preRelease

	return nil
}

func (v Version) String() string               { return v.Original }
func (v Version) MarshalText() ([]byte, error) { return []byte(v.String()), nil }

func (v *Version) UnmarshalText(data []byte) error {
	if string(data) == "" {
		return nil
	}
	return v.Set(string(data))
}

// Compare tests if v is less than, equal to, or greater than versionB,
// returning -1, 0, or +1 respectively.
func (v Version) Compare(versionB Version) int {
	if cmp := recursiveCompare(v.Slice(), versionB.Slice()); cmp != 0 {
		return cmp
	}
	return preReleaseCompare(v, versionB)
}

// Equal tests if v is equal to versionB.
func (v Version) Equal(versionB Version) bool {
	return v.Compare(versionB) == 0
}

// LessThan tests if v is less than versionB.
func (v Version) LessThan(versionB Version) bool {
	return v.Compare(versionB) < 0
}

// Slice converts the comparable parts of the semver into a slice of integers.
func (v Version) Slice() []int64 {
	return v.Patched
}

func (p PreRelease) Slice() []string {
	preRelease := string(p)
	return strings.Split(preRelease, ".")
}

func preReleaseCompare(versionA Version, versionB Version) int {
	a := versionA.PreRelease
	b := versionB.PreRelease

	/* Handle the case where if two versions are otherwise equal it is the
	 * one without a PreRelease that is greater */
	if len(a) == 0 && (len(b) > 0) {
		return 1
	} else if len(b) == 0 && (len(a) > 0) {
		return -1
	}

	// If there is a prerelease, check and compare each part.
	return recursivePreReleaseCompare(a.Slice(), b.Slice())
}

func recursiveCompare(versionA []int64, versionB []int64) int {
	if len(versionA) == 0 && len(versionB) == 0 {
		return 0
	}

	if len(versionA) == 0 {
		versionA = make([]int64, len(versionB))
	} else if len(versionB) == 0 {
		versionB = make([]int64, len(versionA))
	}

	a, b := versionA[0], versionB[0]
	if a > b {
		return 1
	} else if a < b {
		return -1
	}

	return recursiveCompare(versionA[1:], versionB[1:])
}

func recursivePreReleaseCompare(versionA []string, versionB []string) int {
	// A larger set of pre-release fields has a higher precedence than a smaller set,
	// if all of the preceding identifiers are equal.
	if len(versionA) == 0 {
		if len(versionB) > 0 {
			return -1
		}
		return 0
	} else if len(versionB) == 0 {
		// We're longer than versionB so return 1.
		return 1
	}

	a := versionA[0]
	b := versionB[0]

	aInt := false
	bInt := false

	aI, err := strconv.Atoi(versionA[0])
	if err == nil {
		aInt = true
	}

	bI, err := strconv.Atoi(versionB[0])
	if err == nil {
		bInt = true
	}

	// Numeric identifiers always have lower precedence than non-numeric identifiers.
	if aInt && !bInt {
		return -1
	} else if !aInt && bInt {
		return 1
	}

	// Handle Integer Comparison
	if aInt && bInt {
		if aI > bI {
			return 1
		} else if aI < bI {
			return -1
		}
	}

	// Handle String Comparison
	if a > b {
		return 1
	} else if a < b {
		return -1
	}

	return recursivePreReleaseCompare(versionA[1:], versionB[1:])
}

// BumpMajor increments the Major field by 1 and resets all other fields to their default values
func (v *Version) BumpMajor() {
	if len(v.Patched) < 3 {
		v.Patched = append(v.Patched, make([]int64, 3)...)
	}
	v.Patched[0] += 1
	v.Patched[1] = 0
	v.Patched[2] = 0
	v.PreRelease = PreRelease("")
	v.Metadata = ""
}

// BumpMinor increments the Minor field by 1 and resets all other fields to their default values
func (v *Version) BumpMinor() {
	if len(v.Patched) < 3 {
		v.Patched = append(v.Patched, make([]int64, 3)...)
	}
	v.Patched[1] += 1
	v.Patched[2] = 0
	v.PreRelease = PreRelease("")
	v.Metadata = ""
}

// BumpPatch increments the Patch field by 1 and resets all other fields to their default values
func (v *Version) BumpPatch() {
	if len(v.Patched) < 3 {
		v.Patched = append(v.Patched, make([]int64, 3)...)
	}
	v.Patched[2] += 1
	v.PreRelease = PreRelease("")
	v.Metadata = ""
}

// validateIdentifier makes sure the provided identifier satisfies semver spec
func validateIdentifier(id string) error {
	if id != "" && !reIdentifier.MatchString(id) {
		return fmt.Errorf("%s is not a valid semver identifier", id)
	}
	return nil
}

// reIdentifier is a regular expression used to check that pre-release and metadata
// identifiers satisfy the spec requirements
var reIdentifier = regexp.MustCompile(`^[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*$`)

type Versions []*Version

// Sort sorts the given slice of Version
func Sort(versions []*Version) {
	sort.Sort(Versions(versions))
}

func (s Versions) Len() int      { return len(s) }
func (s Versions) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s Versions) Less(i, j int) bool {
	return s[i].LessThan(*s[j])
}

type SemverVersion interface {
	SemverVersion() *Version
}

// Sort Struct giving from SemverVersion function
func SortStruct[S SemverVersion](versions []S) {
	sort.Sort(InterfaceVersion[S](versions))
}

type InterfaceVersion[P SemverVersion] []P

func (s InterfaceVersion[P]) Len() int      { return len(s) }
func (s InterfaceVersion[P]) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s InterfaceVersion[P]) Less(i, j int) bool {
	return s[i].SemverVersion().LessThan(*s[j].SemverVersion())
}
