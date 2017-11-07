package main

import (
	"github.com/jslater89/graviton/data"
)

func main() {
	data.InitMongo("localhost", "graviton")
}
