// Package buildinfo exposes the compiler binding embedded at build time.
package buildinfo

import (
	"encoding/json"
	"fmt"
)

type TypeScript struct {
	Commit         string  `json:"commit"`
	ReleaseVersion *string `json:"releaseVersion"`
}

type Info struct {
	Version    string     `json:"version"`
	TypeScript TypeScript `json:"typescript"`
}

func Current() Info {
	var info Info
	if err := json.Unmarshal([]byte(currentJSON), &info); err != nil {
		panic("invalid generated build info: " + err.Error())
	}
	return info
}

func (info Info) String() string {
	version := ""
	if info.TypeScript.ReleaseVersion != nil {
		version = *info.TypeScript.ReleaseVersion + " "
	}
	return fmt.Sprintf("rslint %s\nTypeScript %s(%s)", info.Version, version, info.TypeScript.Commit)
}
