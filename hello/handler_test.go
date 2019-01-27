package main

import (
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

var testInviteCode = 300 // If you update this number, you still have to find and replace it in the mock objects

var mockInviteCodeFulfillmentRequest = events.APIGatewayProxyRequest{
	Body: `
	{
		"responseId": "69ebe95e-3b44-4954-a2a2-20200418f61a",
		"queryResult": {
		  "queryText": "300",
		  "action": "rsvperwelcome.rsvperwelcome-custom",
		  "parameters": {
			"invite_code": 300
		  },
		  "allRequiredParamsPresent": true,
		  "fulfillmentText": "Thanks! Got your invite code -- 300",
		  "fulfillmentMessages": [
			{
			  "text": {
				"text": [
				  "Thanks! Got your invite code -- 300"
				]
			  }
			}
		  ],
		  "outputContexts": [
			{
			  "name": "projects/rsvper-42ec0/agent/sessions/5c7d44c9-0527-98b1-7402-622458f80cb8/contexts/rsvprinvite-followup",
			  "lifespanCount": 1,
			  "parameters": {
				"invite_code.original": "300",
				"invite_code": 300
			  }
			},
			{
			  "name": "projects/rsvper-42ec0/agent/sessions/5c7d44c9-0527-98b1-7402-622458f80cb8/contexts/rsvperwelcome-followup",
			  "lifespanCount": 1,
			  "parameters": {
				"invite_code.original": "300",
				"invite_code": 300
			  }
			},
			{
			  "name": "projects/rsvper-42ec0/agent/sessions/5c7d44c9-0527-98b1-7402-622458f80cb8/contexts/rsvperwelcome-invitecode-followup",
			  "lifespanCount": 10,
			  "parameters": {
				"invite_code.original": "300",
				"invite_code": 300
			  }
			}
		  ],
		  "intent": {
			"name": "projects/rsvper-42ec0/agent/intents/7a559a2c-5f89-406f-8e3e-d8f4073fedaa",
			"displayName": "rsvper.welcome - invitecode"
		  },
		  "intentDetectionConfidence": 1,
		  "languageCode": "en"
		},
		"originalDetectIntentRequest": {
		  "payload": {}
		},
		"session": "projects/rsvper-42ec0/agent/sessions/5c7d44c9-0527-98b1-7402-622458f80cb8"
	  }`,
}

var mockRsvpFulfillmentRequest = events.APIGatewayProxyRequest{
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

var mockRsvpSlotFillingRequest = events.APIGatewayProxyRequest{
	Body: `
	"responseId": "aec29e25-eb7a-42be-91d8-ea99a7b2221d",
	"queryResult": {
	  "queryText": "rsvp",
	  "action": "rsvperwelcome.rsvperwelcome-custom.rsvperwelcome-custom-custom",
	  "parameters": {
		"vidhi_rsvp": "",
		"garba_rsvp": "",
		"wedding_rsvp": ""
	  },
	  "allRequiredParamsPresent": true,
	  "fulfillmentMessages": [
		{
		  "text": {
			"text": [
			  ""
			]
		  }
		}
	  ],
	  "outputContexts": [
		{
		  "name": "projects/rsvper-42ec0/agent/sessions/5c7d44c9-0527-98b1-7402-622458f80cb8/contexts/rsvprinvite-followup",
		  "parameters": {
			"wedding_rsvp": "",
			"wedding_rsvp.original": "",
			"invite_code.original": "1",
			"garba_rsvp": "",
			"invite_code": 1,
			"garba_rsvp.original": "",
			"vidhi_rsvp": "",
			"vidhi_rsvp.original": ""
		  }
		},
		{
		  "name": "projects/rsvper-42ec0/agent/sessions/5c7d44c9-0527-98b1-7402-622458f80cb8/contexts/rsvperwelcome-followup",
		  "parameters": {
			"wedding_rsvp": "",
			"wedding_rsvp.original": "",
			"invite_code.original": "1",
			"garba_rsvp": "",
			"invite_code": 1,
			"garba_rsvp.original": "",
			"vidhi_rsvp": "",
			"vidhi_rsvp.original": ""
		  }
		},
		{
		  "name": "projects/rsvper-42ec0/agent/sessions/5c7d44c9-0527-98b1-7402-622458f80cb8/contexts/rsvperwelcome-invitecode-followup",
		  "lifespanCount": 9,
		  "parameters": {
			"wedding_rsvp": "",
			"wedding_rsvp.original": "",
			"invite_code.original": "1",
			"garba_rsvp": "",
			"invite_code": 1,
			"garba_rsvp.original": "",
			"vidhi_rsvp": "",
			"vidhi_rsvp.original": ""
		  }
		}
	  ],
	  "intent": {
		"name": "projects/rsvper-42ec0/agent/intents/e9c67536-369b-46cd-9acb-bab986937295",
		"displayName": "rsvper.welcome - invitecode - rsvp"
	  },
	  "intentDetectionConfidence": 1,
	  "languageCode": "en"
	},
	"originalDetectIntentRequest": {
	  "payload": {}
	},
	"session": "projects/rsvper-42ec0/agent/sessions/5c7d44c9-0527-98b1-7402-622458f80cb8"
  }`,
}

var mockWeddingRsvpFulfillmentRequest = events.APIGatewayProxyRequest{
	Body: `
	{
		"responseId": "f6620d32-181b-40c1-be57-1496812fd383",
		"queryResult": {
			"queryText": "2",
			"parameters": {
				"wedding_rsvpd": 2
			},
			"allRequiredParamsPresent": true,
			"fulfillmentMessages": [
				{
					"text": {
						"text": [
							""
						]
					}
				}
			],
			"outputContexts": [
				{
					"name": "projects/rsvper-42ec0/agent/sessions/7dc551fb-1701-460b-c079-d4abbabda913/contexts/rsvprinvite-followup",
					"lifespanCount": 5,
					"parameters": {
						"invite_code.original": "300",
						"wedding_rsvpd.original": "2",
						"invite_code": 300,
						"wedding_rsvpd": 2
					}
				},
				{
					"name": "projects/rsvper-42ec0/agent/sessions/7dc551fb-1701-460b-c079-d4abbabda913/contexts/rsvperwelcome-followup",
					"parameters": {
						"invite_code.original": "300",
						"wedding_rsvpd.original": "2",
						"invite_code": 300,
						"wedding_rsvpd": 2
					}
				},
				{
					"name": "projects/rsvper-42ec0/agent/sessions/7dc551fb-1701-460b-c079-d4abbabda913/contexts/rsvperwelcome-invitecode-followup",
					"lifespanCount": 9,
					"parameters": {
						"invite_code.original": "300",
						"wedding_rsvpd.original": "2",
						"invite_code": 300,
						"wedding_rsvpd": 2
					}
				}
			],
			"intent": {
				"name": "projects/rsvper-42ec0/agent/intents/31a53897-a029-4732-89ca-ad93cc64e06c",
				"displayName": "rsvper.rsvp-wedding"
			},
			"intentDetectionConfidence": 1,
			"languageCode": "en"
		},
		"originalDetectIntentRequest": {},
		"session": "projects/rsvper-42ec0/agent/sessions/7dc551fb-1701-460b-c079-d4abbabda913"
	}
	`,
}

func TestInviteCodeFulfillmentHandler(t *testing.T) {
	response, err := Handler(mockWeddingRsvpFulfillmentRequest)
	if err != nil {
		t.Errorf("Error: +%v", err)
	}

	t.Logf("Response: +%v", response)

	t.Error()
}

// func TestInviteCodeFulfillmentHandler(t *testing.T) {
// 	response, err := Handler(mockInviteCodeFulfillmentRequest)
// 	if err != nil {
// 		t.Errorf("Error: +%v", err)
// 	}

// 	t.Logf("Response: +%v", response)
// }

// func BenchmarkInviteCodeFulfillmentHandler(b *testing.B) {
// 	response, err := Handler(mockInviteCodeFulfillmentRequest)
// 	if err != nil {
// 		b.Errorf("Error: +%v", err)
// 	}

// 	b.Logf("Response: +%v", response)
// }

// func TestRsvpFulfillmentHandler(t *testing.T) {
// 	response, err := Handler(mockRsvpFulfillmentRequest)
// 	if err != nil {
// 		t.Errorf("Error: +%v", err)
// 	}

// 	t.Logf("Response: +%v", response)
// }

// func BenchmarkRsvpFulfillmentHandler(b *testing.B) {
// 	response, err := Handler(mockRsvpFulfillmentRequest)
// 	if err != nil {
// 		b.Errorf("Error: +%v", err)
// 	}

// 	b.Logf("Response: +%v", response)
// }
// func TestRsvpSlotFillingHandler(t *testing.T) {
// 	response, err := Handler(mockRsvpSlotFillingRequest)
// 	if err != nil {
// 		t.Errorf("Error: +%v", err)
// 	}

// 	t.Logf("Response: +%v", response)
// }

// func BenchmarkRsvpSlotFillingHandler(b *testing.B) {
// 	response, err := Handler(mockRsvpSlotFillingRequest)
// 	if err != nil {
// 		b.Errorf("Error: +%v", err)
// 	}

// 	b.Logf("Response: +%v", response)
// }
// func TestEnvVariableExists(t *testing.T) {
// 	googleAPICreds := os.Getenv("GOOGLE_API_CREDS")
// 	if googleAPICreds == "" {
// 		t.Error("Couldn't find the GOOGLE_API_CREDS environment variable")
// 	}
// }

// // Can't test because starting up the program runs main which runs the Handler and
// // func TestSearchForInvitedFamily(t *testing.T) {
// // 	family, err := SearchForInvitedFamily(13)
// // 	if err != nil {
// // 		t.Errorf("Error: +%v", err)
// // 	}

// // 	t.Logf("Family: +%v", family)
// // }
