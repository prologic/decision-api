package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/flagship-io/decision-api/internal/apilogic"
	"github.com/flagship-io/decision-api/internal/handle"
	"github.com/flagship-io/decision-api/internal/models"
	"github.com/flagship-io/decision-api/internal/utils"
	"github.com/flagship-io/flagship-proto/decision_response"
	"github.com/golang/protobuf/jsonpb"
	"gitlab.com/canarybay/protobuf/ptypes.git/flags"
)

// FlagMetadata represents the metadata informations about a flag key
type FlagMetadata struct {
	CampaignID       string `json:"campaignId"`
	VariationGroupID string `json:"variationGroupId"`
	VariationID      string `json:"variationId"`
}

// FlagInfos represents the list of flags
type FlagInfos struct {
	Flags map[string]FlagInfo `json:"flags"`
}

// FlagInfo represents the informations about a flag key
type FlagInfo struct {
	Value    interface{}  `json:"value"`
	Metadata FlagMetadata `json:"metadata"`
}

func Flags(context *models.DecisionContext) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		apilogic.HandleCampaigns(w, req, context, requestFlagsHandler, utils.NewTracker())
	}
}

func requestFlagsHandler(w http.ResponseWriter, handleRequest *handle.Request, err error) {
	if err != nil {
		utils.WriteClientError(w, http.StatusBadRequest, err.Error())
		return
	}

	sendFlagsResponse(w, handleRequest.DecisionResponse)
}

func sendFlagsResponse(w http.ResponseWriter, decisionResponse *decision_response.DecisionResponse) {
	ma := jsonpb.Marshaler{EmitDefaults: true}

	flagInfos := flags.FlagInfos{
		Flags: make(map[string]*flags.FlagInfo),
	}

	for _, c := range decisionResponse.Campaigns {
		if c.GetVariation() != nil && c.GetVariation().GetModifications() != nil && c.GetVariation().GetModifications().GetValue() != nil && c.GetVariation().GetModifications().GetValue().GetFields() != nil {
			for k, v := range c.GetVariation().GetModifications().GetValue().GetFields() {
				_, exists := flagInfos.Flags[k]
				if exists {
					continue
				}
				flagInfos.Flags[k] = &flags.FlagInfo{
					Value: v,
					Metadata: &flags.FlagMetadata{
						CampaignId:        c.GetId().Value,
						Variation_GroupId: c.GetVariationGroupId().Value,
						VariationId:       c.GetVariation().GetId().Value,
					},
				}
			}
		}
	}

	flagInfosJSON := FlagInfos{}
	flagString, err := ma.MarshalToString(&flagInfos)

	if err != nil {
		utils.WriteServerError(w, err)
		return
	}

	err = json.Unmarshal([]byte(flagString), &flagInfosJSON)
	if err != nil {
		utils.WriteServerError(w, err)
		return
	}

	utils.WriteJSONOk(w, flagInfosJSON.Flags)
}
