package version

import (
	"fmt"
	"strconv"
	"strings"
)

type parsedVersion struct {
	major          int
	minor          int
	patch          int
	prerelease     string
	prereleaseNum  int
	hasPrerelease  bool
	hasPrereleaseN bool
}

// Normalize strips a leading "v" and surrounding whitespace from version strings.
func Normalize(v string) string {
	v = strings.TrimSpace(v)
	if strings.HasPrefix(strings.ToLower(v), "v") {
		return v[1:]
	}
	return v
}

// Compare returns -1 if a < b, 0 if equal, and 1 if a > b.
func Compare(a, b string) int {
	pa, errA := parse(Normalize(a))
	pb, errB := parse(Normalize(b))
	if errA != nil || errB != nil {
		return strings.Compare(Normalize(a), Normalize(b))
	}

	if pa.major != pb.major {
		return cmpInt(pa.major, pb.major)
	}
	if pa.minor != pb.minor {
		return cmpInt(pa.minor, pb.minor)
	}
	if pa.patch != pb.patch {
		return cmpInt(pa.patch, pb.patch)
	}

	if !pa.hasPrerelease && !pb.hasPrerelease {
		return 0
	}
	if !pa.hasPrerelease && pb.hasPrerelease {
		return 1
	}
	if pa.hasPrerelease && !pb.hasPrerelease {
		return -1
	}

	if pa.prerelease != pb.prerelease {
		return strings.Compare(pa.prerelease, pb.prerelease)
	}
	if pa.hasPrereleaseN && pb.hasPrereleaseN {
		return cmpInt(pa.prereleaseNum, pb.prereleaseNum)
	}
	if pa.hasPrereleaseN && !pb.hasPrereleaseN {
		return 1
	}
	if !pa.hasPrereleaseN && pb.hasPrereleaseN {
		return -1
	}
	return 0
}

// IsPrerelease reports whether v looks like a dev/beta prerelease.
func IsPrerelease(v string) bool {
	p, err := parse(Normalize(v))
	if err != nil {
		return strings.Contains(strings.ToLower(Normalize(v)), "beta")
	}
	return p.hasPrerelease
}

func parse(v string) (parsedVersion, error) {
	if v == "" || v == "dev" {
		return parsedVersion{}, fmt.Errorf("unversioned")
	}

	base := v
	prerelease := ""
	if idx := strings.Index(v, "-"); idx >= 0 {
		base = v[:idx]
		prerelease = v[idx+1:]
	}

	parts := strings.Split(base, ".")
	if len(parts) != 3 {
		return parsedVersion{}, fmt.Errorf("invalid version: %s", v)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return parsedVersion{}, err
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return parsedVersion{}, err
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return parsedVersion{}, err
	}

	out := parsedVersion{
		major: major,
		minor: minor,
		patch: patch,
	}
	if prerelease != "" {
		out.hasPrerelease = true
		if dot := strings.LastIndex(prerelease, "."); dot >= 0 {
			if n, err := strconv.Atoi(prerelease[dot+1:]); err == nil {
				out.prerelease = prerelease[:dot]
				out.prereleaseNum = n
				out.hasPrereleaseN = true
				return out, nil
			}
		}
		out.prerelease = prerelease
	}
	return out, nil
}

func cmpInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}
