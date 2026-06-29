package version

import (
	"fmt"
	"os"
	"text/template"
)

// ╦ ╦┌─┐─┐ ┬┌─┐╔═╗┌─┐
// ╠═╣├┤ ┌┴┬┘├─┤║ ╦│ │
// ╩ ╩└─┘┴ └─┴ ┴╚═╝└─┘

// ┃ ┃┏━┛┃ ┃┏━┃┏━┛┏━┃
// ┏━┃┏━┛ ┛ ┏━┃┃ ┃┃ ┃
// ┛ ┛━━┛┛ ┛┛ ┛━━┛━━┛

// ░█░█░█▀▀░█░█░█▀█░█▀▀░█▀█
// ░█▀█░█▀▀░▄▀▄░█▀█░█░█░█░█
// ░▀░▀░▀▀▀░▀░▀░▀░▀░▀▀▀░▀▀▀

//  +-+-+-+-+-+-+
//  |H|e|x|a|G|o|
//  +-+-+-+-+-+-+

func Splash() {

	var (
		splashTemplate = `
┓┏      ┏┓    Version: {{ .Major }}.{{ .Minor }}.{{ .Patch }}{{ if .Extra  }}-{{ .Extra }}{{ end }}
┣┫┏┓┓┏┏┓┃┓┏┓  Build: {{ .BuildDate }}
┛┗┗ ┛┗┗┻┗┛┗┛  Commit: {{ .Commit }}

`
	)

	t, err := template.New("splash").Parse(splashTemplate)
	if err != nil {
		fmt.Printf("Error parsing template: %+v", err)
	}

	if err := t.Execute(os.Stdout, CurrentVersion()); err != nil {
		fmt.Printf("Error executing template: %+v", err)
	}
}
