package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func onfig() {
	file, _ := os.Open("config.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println(configuration) // output: [UserA, UserB]
}
