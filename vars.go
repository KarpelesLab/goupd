package goupd

import (
	"path"
	"runtime/debug"
)

var (
	PROJECT_NAME string = "unconfigured"
	MODE         string = "DEV"
	CHANNEL      string = "stable"
	GIT_TAG      string = ""
	DATE_TAG     string = "0"
	HOST         string = "https://dist-go.tristandev.net/"
)

func init() {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	if PROJECT_NAME == "unconfigured" {
		PROJECT_NAME = path.Base(bi.Path) // full path of project
	}
	for _, setting := range bi.Settings {
		switch setting.Key {
		case "vcs.revision":
			// {Key:vcs.revision Value:dfb3603f8cb1fa40d5c3a5b9bbbdf6d316e7e2fe}
			// check length just in case so we don't cause panic
			if GIT_TAG == "" && len(setting.Value) >= 7 {
				GIT_TAG = setting.Value[:7]
			}
		case "vcs.time":
			if DATE_TAG != "0" {
				break
			}
			// {Key:vcs.time Value:2022-06-01T01:55:46Z}
			v := make([]byte, 0, 14) // normal length
			for _, c := range setting.Value {
				if c >= '0' && c <= '9' {
					v = append(v, byte(c))
				}
			}
			if len(v) == 14 {
				DATE_TAG = string(v)
			}
		}
	}
}

func GetVars() map[string]string {
	return map[string]string{
		"PROJECT_NAME": PROJECT_NAME,
		"MODE":         MODE,
		"CHANNEL":      CHANNEL,
		"GIT_TAG":      GIT_TAG,
		"DATE_TAG":     DATE_TAG,
		"HOST":         HOST,
	}
}
