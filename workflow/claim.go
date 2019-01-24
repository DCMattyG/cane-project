package workflow

import (
	"cane-project/database"
	"cane-project/model"
	"fmt"

	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/rs/xid"
)

// Claim Alias
type Claim model.WorkflowClaim

// GenerateClaim Function
func GenerateClaim() model.WorkflowClaim {
	var claim model.WorkflowClaim

	claim.ClaimCode = xid.New()

	return claim
}

// ---> Finish by making this Save & Update

// SaveClaim Function
func (c *Claim) SaveClaim() {
	filter := primitive.M{
		"claimCode": c.ClaimCode,
	}

	_, findErr := database.FindOne("workflow", "claims", filter)

	if findErr == nil {
		fmt.Println(findErr)
		return
	}

	_, saveErr := database.Save("workflow", "claims", c)

	if saveErr != nil {
		fmt.Println(saveErr)
		return
	}

	return
}
