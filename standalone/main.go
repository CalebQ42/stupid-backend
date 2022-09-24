package main

import "flag"

func main() {
	appList := flag.String("apps", "testing", "Comma deliniated list of apps to use. If none given, uses \"testing\". If Collections are not cre")
	port := flag.Int("port", 4223, "Port to open requests on. Defaults to 4223.")
	keysDir := flag.String("tlsdir", "", "Directory with key.pem and cert.pem. Defaults to $HOME. Required.")
	flag.Parse()
	if *keysDir == "" {

	}
}
