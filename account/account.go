package account

import (
	"cane/database"
	"fmt"
	"net"

	"github.com/mitchellh/mapstructure"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

// DeviceAccount Struct
type DeviceAccount struct {
	ID       primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name     string             `json:"name" bson:"name"`
	IP       net.IP             `json:"ip" bson:"ip"`
	UserName string             `json:"username" bson:"username"`
	Password string             `json:"password" bson:"password"`
	AuthType string             `json:"authtype" bson:"authtype"`
}

// SaveAccount Function
func SaveAccount(account DeviceAccount) {
	database.SelectDatabase("account", "device")

	accountID := database.InsertToDB(account)

	fmt.Print("Inserted ID: ")
	fmt.Println(accountID)

	return
}

// LoadAccount Function
func LoadAccount(accountName string) DeviceAccount {
	var foundAccount DeviceAccount

	database.SelectDatabase("account", "device")

	filter := primitive.M{"name": accountName}
	result := database.FindOneInDB(filter)
	mapstructure.Decode(result, &foundAccount)

	return foundAccount
}
