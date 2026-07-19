package main

import "github.com/mymmrac/telego"

func main() {
	_, _ = telego.NewBot("1:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", telego.WithDiscardLogger())
}
