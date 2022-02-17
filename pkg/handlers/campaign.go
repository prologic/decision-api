package handlers

import (
	"net/http"
	"strconv"

	"github.com/flagship-io/decision-api/internal/apilogic"
	"github.com/flagship-io/decision-api/internal/handle"
	"github.com/flagship-io/decision-api/internal/utils"
	"github.com/flagship-io/decision-api/pkg/connectors"
	"github.com/flagship-io/flagship-proto/decision_response"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

func Campaign(context *connectors.DecisionContext) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		apilogic.HandleCampaigns(w, req, context, requestCampaignHandler, utils.NewTracker())
	}
}

func requestCampaignHandler(w http.ResponseWriter, handleRequest *handle.Request, err error) {
	if err != nil {
		utils.WriteClientError(w, http.StatusBadRequest, err.Error())
		return
	}

	if handleRequest.DecisionRequest.GetFormatResponse() != nil && handleRequest.DecisionRequest.GetFormatResponse().GetValue() {
		sendSingleFormatResponse(w, handleRequest.DecisionResponse.Campaigns[0])
		return
	}

	if len(handleRequest.DecisionResponse.Campaigns) == 0 {
		utils.WriteNoContent(w)
		return
	}

	sendSingleResponse(w, handleRequest.DecisionResponse.Campaigns[0])
}

func sendSingleResponse(w http.ResponseWriter, campaignDecisionResponse *decision_response.Campaign) {
	data, err := protojson.Marshal(campaignDecisionResponse)
	if err != nil {
		utils.WriteServerError(w, err)
		return
	}
	w.Write(data)
}

func sendSingleFormatResponse(w http.ResponseWriter, campaignDecisionResponse *decision_response.Campaign) {
	contentType := "application/json"
	switch campaignDecisionResponse.GetVariation().GetModifications().GetType() {
	case decision_response.ModificationsType_IMAGE:
		contentType = "image"
	case decision_response.ModificationsType_TEXT:
		contentType = "text/plain"
	case decision_response.ModificationsType_HTML:
		contentType = "text/html"
	default:
		sendSingleResponse(w, campaignDecisionResponse)
		return
	}

	fields := campaignDecisionResponse.GetVariation().GetModifications().GetValue().GetFields()
	var value *structpb.Value
	for _, v := range fields {
		value = v
	}

	if value == nil {
		sendSingleResponse(w, campaignDecisionResponse)
		return
	}

	dataValue := ""
	switch value.Kind.(type) {
	case (*structpb.Value_StringValue):
		dataValue = value.GetStringValue()
	case (*structpb.Value_BoolValue):
		dataValue = strconv.FormatBool(value.GetBoolValue())
	case (*structpb.Value_NumberValue):
		dataValue = strconv.FormatFloat(value.GetNumberValue(), 'E', -1, 64)
	}

	w.Header().Add("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(dataValue))
}
