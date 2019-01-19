package main

import (
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

var mockRequest = events.APIGatewayProxyRequest{
	Body: `
	{
		"responseId":"96b07e3f-2186-491d-863a-caddfc7fd2bd",
		"queryResult":{
		   "queryText":"8",
		   "action":"input.rsvp",
		   "parameters":{
			  "weddingRsvpdInvitees":8,
			  "vidhiRsvpdInvitees":6,
			  "garbaRsvpdInvitees":7
		   },
		   "allRequiredParamsPresent":true,
		   "fulfillmentText":"Great -- Ive got your rsvps down. See you at the wedding!! <3",
		   "fulfillmentMessages":[
			  {
				 "text":{
					"text":[
					   "Great -- Ive got your rsvps down. See you at the wedding!! <3"
					]
				 }
			  }
		   ],
		   "outputContexts":[
			  {
				 "name":"projects/rsvper-42ec0/agent/sessions/e22972e8-1cc2-1556-3df6-22d6316a815f/contexts/rsvp_context",
				 "lifespanCount":25,
				 "parameters":{
					"vidhiRsvpdInvitees":6,
					"weddingRsvpdInvitees":8,
					"vidhiRsvpdInvitees.original":"6",
					"weddingRsvpdInvitees.original":"8",
					"garbaRsvpdInvitees.original":"7",
					"garbaRsvpdInvitees":7
				 }
			  }
		   ],
		   "intent":{
			  "name":"projects/rsvper-42ec0/agent/intents/a919510e-d7b9-43e9-82d6-8bdf57ddcc85",
			  "displayName":"rsvper.rsvp",
			  "endInteraction":true
		   },
		   "intentDetectionConfidence":1,
		   "diagnosticInfo":{
			  "end_conversation":true
		   },
		   "languageCode":"en"
		},
		"originalDetectIntentRequest":{
		   "payload":{
	 
		   }
		},
		"session":"projects/rsvper-42ec0/agent/sessions/e22972e8-1cc2-1556-3df6-22d6316a815f"
	 }`,
}

func TestHandler(t *testing.T) {
	response, err := Handler(mockRequest)
	if err != nil {
		t.Errorf("Error: +%v", err)
	}

	t.Logf("Response: +%v", response)
}

func BenchmarkHandler(b *testing.B) {
	response, err := Handler(mockRequest)
	if err != nil {
		b.Errorf("Error: +%v", err)
	}

	b.Logf("Response: +%v", response)
}

// Can't test because starting up the program runs main which runs the Handler and
// func TestSearchForInvitedFamily(t *testing.T) {
// 	family, err := SearchForInvitedFamily(13)
// 	if err != nil {
// 		t.Errorf("Error: +%v", err)
// 	}

// 	t.Logf("Family: +%v", family)
// }
