package account

import (
	"net"

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
