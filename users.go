package main

import (
	"encoding/json"
	"fmt"
)

func (m Users) Restore(row []byte) {
	var usr User
	err := json.Unmarshal(row, &usr)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return
	}
	users[usr.Name] = usr
}
