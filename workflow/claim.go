package workflow

import (
	"cane-project/database"
	"cane-project/model"
	"fmt"

	"github.com/mitchellh/mapstructure"

	"github.com/fatih/structs"

	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/segmentio/ksuid"
)

// Claim Alias
type Claim model.WorkflowClaim

// GenerateClaim Function
func GenerateClaim() Claim {
	var claim Claim

	claim.ClaimCode = ksuid.New().String()
	claim.WorkflowResults = make(map[string]model.StepResult)

	return claim
}

// Save Function
func (c *Claim) Save() {
	var replace primitive.M

	filter := primitive.M{
		"claimCode": c.ClaimCode,
	}

	replace = structs.Map(c)

	delete(replace, "_id")

	replaceVal, replaceErr := database.FindAndReplace("workflows", "claims", filter, replace)

	if replaceErr != nil {
		fmt.Println(replaceErr)
		return
	}

	mapstructure.Decode(replaceVal, &c)
}
