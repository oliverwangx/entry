package main

import (
	"encoding/json"
	"fmt"
	"shopee-backend-entry-task/model"
)

func main() {
	user := &model.User{}
	val := `{"Username":"oliver","Password":"1291ba28105cd226e2c12a436236e3f4","Nickname":"super_fancy","Avatar":"image/person4.png"}`
	err := json.Unmarshal([]byte(val), user)
	fmt.Println(err)
}
