package version

import (
	"fmt"
	"io"
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

const splashTemplate = `
┓┏      ┏┓    Version: {{ .Major }}.{{ .Minor }}.{{ .Patch }}{{ if .Extra  }}-{{ .Extra }}{{ end }}
┣┫┏┓┓┏┏┓┃┓┏┓  Build: {{ .BuildDate }}
┛┗┗ ┛┗┗┻┗┛┗┛  Commit: {{ .Commit }}

`

func Splash(stdWriter, errWriter io.Writer) {
	t, err := template.New("splash").Parse(splashTemplate)
	if err != nil {
		fmt.Fprintf(errWriter, "Error parsing template: %+v", err)
		return
	}

	if err := t.Execute(stdWriter, CurrentVersion()); err != nil {
		fmt.Fprintf(errWriter, "Error executing template: %+v", err)
	}
}
