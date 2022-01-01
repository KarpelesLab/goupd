package goupd

var (
	PROJECT_NAME string = "unconfigured"
	MODE         string = "DEV"
	CHANNEL      string = "stable"
	GIT_TAG      string = ""
	DATE_TAG     string = "0"
	HOST         string = "https://dist-go.tristandev.net/"
)

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
