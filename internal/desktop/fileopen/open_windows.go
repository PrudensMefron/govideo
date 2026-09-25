package fileopen

func command(path string) (string, []string) {
	return "rundll32.exe", []string{"url.dll,FileProtocolHandler", path}
}
