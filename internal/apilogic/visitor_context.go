package apilogic

import (
	"github.com/flagship-io/decision-api/internal/handle"
	"github.com/flagship-io/decision-api/pkg/connectors"
	"github.com/flagship-io/decision-api/pkg/models"
)

// SendVisitorContext sends a pubsub event to handle visitor context
func SendVisitorContext(handleRequest *handle.Request) {
	// If sendContext explicitely set to !true or context is empty, return
	if !handleRequest.SendContextEvent || len(handleRequest.DecisionRequest.Context) == 0 {
		return
	}

	contextMap := map[string]interface{}{}
	for k, v := range handleRequest.DecisionRequest.Context {
		contextMap[k] = v.AsInterface()
	}

	visitorContext := &models.VisitorContext{
		EnvID:     handleRequest.DecisionContext.EnvID,
		VisitorID: handleRequest.DecisionRequest.VisitorId.Value,
		Context:   contextMap,
		Timestamp: handleRequest.Time.UnixNano() / 1000000,
	}

	err := handleRequest.DecisionContext.HitProcessor.TrackHits(connectors.TrackingHits{VisitorContext: []*models.VisitorContext{
		visitorContext,
	}})
	if err != nil {
		handleRequest.Logger.Errorf("Error on queuing visitor context : %v", err)
	}
}
