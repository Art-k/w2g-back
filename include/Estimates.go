package include

import "net/http"

type EstimateUnit struct {
	Model
	OwnerUserID     string
	Name            string
	Position        int
	NumberOfSeconds float32
}

func APIEstimate(w http.ResponseWriter, r *http.Request) {

	var user User
	user, err := GetUserByID(r.Context().Value("user").(string))
	if err != nil {
		ResponseInternalServerError(w)
	}

	switch r.Method {
	case "GET":

		var recs []EstimateUnit
		Db.Where("owner_user_id = ?", user.ID).Order("position asc").Find(&recs)
		response := PrepareHTTPResponse(&recs, len(recs), len(recs))
		ResponseOK(w, response)

	default:
		ResponseUnknown(w, "Method is not allowed")
	}
}
