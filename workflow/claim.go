package workflow

import (
	"cane-project/database"
	"cane-project/model"
	"cane-project/util"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
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

// LoadClaim Function
func LoadClaim(w http.ResponseWriter, r *http.Request) {
	claimCode := chi.URLParam(r, "claim")

	foundVal, foundErr := GetClaimFromDB(claimCode)

	if foundErr != nil {
		fmt.Println(foundErr)
		util.RespondWithError(w, http.StatusBadRequest, "claim code not found")
		return
	}

	util.RespondwithJSON(w, http.StatusOK, foundVal)
}

// GetClaimFromDB Function
func GetClaimFromDB(claimCode string) (model.WorkflowClaim, error) {
	var claim model.WorkflowClaim

	fmt.Println("ClaimCode: ", claimCode)

	filter := primitive.M{
		"claimCode": claimCode,
	}

	foundVal, foundErr := database.FindOne("workflows", "claims", filter)

	if foundErr != nil {
		fmt.Println(foundErr)
		return claim, foundErr
	}

	mapErr := mapstructure.Decode(foundVal, &claim)

	if mapErr != nil {
		fmt.Println(mapErr)
		return claim, mapErr
	}

	return claim, nil
}
