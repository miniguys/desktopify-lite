package app

import (
	"fmt"
	"strings"
)

const embeddedLogo = ` ____            _    _              _  __       
 |  _ \  ___  ___| | _| |_ ___  _ __ (_)/ _|_   _ 
 | | | |/ _ \/ __| |/ / __/ _ \| '_ \| | |_| | | |
 | |_| |  __/\__ \   <| || (_) | |_) | |  _| |_| |
 |____/ \___||___/_|\_\\__\___/| .__/|_|_|  \__, |
  miniguys team                |_|lite      |___/
`

func printLogo() {
	logo := strings.TrimRight(embeddedLogo, "\r\n")
	if strings.TrimSpace(logo) == "" {
		return
	}

	renderedLogo := styleBorder.Render(logo)

	fmt.Print(renderedLogo)
	fmt.Println()
}
