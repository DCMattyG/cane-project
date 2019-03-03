package workflow

import (
	"cane-project/database"
	"cane-project/model"
	"cane-project/util"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/mitchellh/mapstructure"

	"github.com/fatih/structs"

	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/mongodb/mongo-go-driver/mongo/options"
	"github.com/segmentio/ksuid"
)

// Claim Alias
type Claim model.WorkflowClaim

// GenerateClaim Function
func GenerateClaim() Claim {
	var claim Claim

	claim.ClaimCode = ksuid.New().String()
	claim.WorkflowResults = make(map[string]model.StepResult)
	claim.Timestamp = time.Now().String()
	claim.CurrentStatus = 0

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

// GetClaims Function
func GetClaims(w http.ResponseWriter, r *http.Request) {
	var opts options.FindOptions
	var claims []string

	foundVal, foundErr := database.FindAll("workflows", "claims", primitive.M{}, opts)

	if foundErr != nil {
		util.RespondWithError(w, http.StatusBadRequest, "no claims found")
		return
	}

	if len(foundVal) == 0 {
		util.RespondWithError(w, http.StatusBadRequest, "empty claims list")
		return
	}

	for key := range foundVal {
		claims = append(claims, foundVal[key]["claimCode"].(string))
	}

	util.RespondwithJSON(w, http.StatusOK, map[string][]string{"claims": claims})
}

// GetClaim Function
func GetClaim(w http.ResponseWriter, r *http.Request) {
	claimCode := chi.URLParam(r, "claimcode")

	foundVal, foundErr := GetClaimFromDB(claimCode)

	if foundErr != nil {
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
		return claim, foundErr
	}

	mapErr := mapstructure.Decode(foundVal, &claim)

	if mapErr != nil {
		return claim, mapErr
	}

	return claim, nil
}
