package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes/struct"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/golang/protobuf/jsonpb"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	sheets "google.golang.org/api/sheets/v4"
	dialogflow "google.golang.org/genproto/googleapis/cloud/dialogflow/v2"
)

// Response is of type APIGatewayProxyResponse since we're leveraging the
// AWS Lambda Proxy Request functionality (default behavior)
//
// https://serverless.com/framework/docs/providers/aws/events/apigateway/#lambda-proxy-integration
type Response events.APIGatewayProxyResponse
type Event struct {
	Name             string
	InvitedCol       string
	RsvpdCol         string
	DialogflowAction string
}

const (
	INVITED_FAMILY       = "INVITED_FAMILY"
	UPDATE_EVENT         = "UPDATE_EVENT"
	TOTAL_INVITED_FAMILY = 9999
	SPREADSHEET_ID       = "1FJPePAwh8Xy9revrg8-ANn7GK2Xwd0Xe_6DdLqDujbc"
	MAX_INVITEES         = 9999
	NULL_INVITEES        = -1
)

type InvitedFamily struct {
	Origin         string
	Name           string
	InviteName     string
	InviteCode     int
	VidhiInvited   int
	VidhiRsvpd     int
	GarbaInvited   int
	GarbaRsvpd     int
	WeddingInvited int
	WeddingRsvpd   int
}

func (invitedFamily *InvitedFamily) totalEventsInvitedTo() int {
	var totalEvents int
	switch{
	case invitedFamily.VidhiInvited > 0:
		totalEvents++
	case invitedFamily.GarbaInvited > 0:
		totalEvents++
	case invitedFamily.WeddingInvited > 0:
		totalEvents++
	}
	return totalEvents
}

var Vidhi = Event{Name: "VIDHI", InvitedCol: "E", RsvpdCol: "F", DialogflowAction: "actions_rsvp_vidhi"}
var Garba = Event{Name: "GARBA", InvitedCol: "G", RsvpdCol: "H", DialogflowAction: "actions_rsvp_garba"}
var Wedding = Event{Name: "WEDDING", InvitedCol: "I", RsvpdCol: "J", DialogflowAction: "actions_rsvp_wedding"}

var sessionID string
var responseID string
var request string

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Println("Received body: ", request.Body)

	var err error
	wr := dialogflow.WebhookRequest{}
	unmarshaller := &jsonpb.Unmarshaler{AllowUnknownFields: true}
	if err = unmarshaller.Unmarshal(strings.NewReader(request.Body), &wr); err != nil {
		log.Fatal(err)
	}
	responseID = wr.ResponseId
	sessionID = wr.Session
	// request = fmt.Sprintf("%+v", request.Body)
	log.Printf("Processing responseId: %s and sessionId: %s", responseID, sessionID)
	// log.Printf("Parsed body: +%v", wr.QueryResult.OutputContexts[0].Parameters)
	intent := wr.QueryResult.Intent.DisplayName
	var message string
	var followupIntentName string
	switch intent {
	case "rsvper.welcome":
		followupIntentName = "rsvper.welcome-invitecode"
	case "rsvper.welcome - invitecode":
		// Given invite code return number of invitees
		fields := wr.QueryResult.Parameters.Fields
		inviteCode := int(fields["invite_code"].GetNumberValue())
		log.Printf("\nIntent: %s - Starting fulfillment for invite code: %d", intent, inviteCode)
		message, _ = InviteCodeFulfillment(inviteCode)
	case "rsvper.welcome - invitecode - yes":
		// Given invite code return number of invitees
		inviteCode := getInviteCodeFromContext(wr.QueryResult.OutputContexts)
		log.Printf("\nIntent: %s - Starting fulfillment for invite code: %d", intent, inviteCode)
		_, followupIntentName = InviteCodeFulfillment(inviteCode)	
	case "rsvper.rsvp-wedding":
		fallthrough
	case "rsvper.welcome - invitecode - yes - wedding":
		fallthrough
	case "rsvper.welcome - invitecode - wedding":
		// Return which event values have to be filled & save updates
		event := Wedding
		phoneNumber := getPhoneNumberFromContext(wr.QueryResult.OutputContexts)
		inviteCode := getInviteCodeFromContext(wr.QueryResult.OutputContexts)
		if inviteCode == -1 {
			log.Fatalf("%s | %s | Couldn't find the invite code", sessionID, responseID)
		}
		log.Printf("\nIntent: %s - Starting fulfillment for invite code: %d", intent, inviteCode)
		followupIntentName = saveRsvpCnt(event, wr.QueryResult.OutputContexts[0].Parameters.Fields, inviteCode, phoneNumber)
	case "rsvper.welcome - invitecode - yes - garba":
		fallthrough
	case "rsvper.welcome - invitecode - garba":
		// Return which event values have to be filled & save updates
		event := Garba
		phoneNumber := getPhoneNumberFromContext(wr.QueryResult.OutputContexts)
		inviteCode := getInviteCodeFromContext(wr.QueryResult.OutputContexts)
		if inviteCode == -1 {
			log.Fatalf("%s | %s | Couldn't find the invite code", sessionID, responseID)
		}
		log.Printf("\nIntent: %s - Starting fulfillment for invite code: %d", intent, inviteCode)
		followupIntentName = saveRsvpCnt(event, wr.QueryResult.OutputContexts[0].Parameters.Fields, inviteCode, phoneNumber)
	case "rsvper.welcome - invitecode - yes - vidhi":
		fallthrough
	case "rsvper.welcome - invitecode - vidhi":
		// Return which event values have to be filled & save updates
		event := Vidhi
		phoneNumber := getPhoneNumberFromContext(wr.QueryResult.OutputContexts)
		inviteCode := getInviteCodeFromContext(wr.QueryResult.OutputContexts)
		if inviteCode == -1 {
			log.Fatalf("%s | %s | Couldn't find the invite code", sessionID, responseID)
		}
		log.Printf("\nIntent: %s - Starting fulfillment for invite code: %d", intent, inviteCode)
		followupIntentName = saveRsvpCnt(event, wr.QueryResult.OutputContexts[0].Parameters.Fields, inviteCode, phoneNumber)
	default:
		log.Printf("\nNo slot-filling or fulfillment functions matched for intent: %s", wr.QueryResult.Intent.DisplayName)
	}

	respBody := createDialogflowResponse(message, followupIntentName)
	return events.APIGatewayProxyResponse{Body: respBody, StatusCode: 200}, nil
}

func createDialogflowResponse(message string, followupIntentName string) string {
	// TODO: fill out the rest of the fields
	responseBody := dialogflow.WebhookResponse{}

	if message != "" {
		responseBody.FulfillmentText = message
	}

	if followupIntentName != "" {
		followupIntent := dialogflow.EventInput{
			LanguageCode: "en",
			Name:         followupIntentName,
		}
		log.Printf("Followup Intent: %+v", followupIntent)
		responseBody.FollowupEventInput = &followupIntent
	}

	log.Printf("\nResponse body: %+v", responseBody)

	var buf bytes.Buffer

	body, err := json.Marshal(responseBody)
	if err != nil {
		log.Fatalf("Unable to parse error response - error: ", err)
		// return events.APIGatewayProxyResponse{StatusCode: 400}, err
	}
	json.HTMLEscape(&buf, body)

	return buf.String()
}

func saveRsvpCnt(event Event, parameters map[string]*structpb.Value, inviteCode int, phoneNumber string) string {
	rsvpCnt := getRsvpCounts(event, parameters)
	if rsvpCnt == -1 {
		log.Fatalf("%s | %s | Couldn't find the rsvp count for event: %s", sessionID, responseID, event.Name)
	}

	eventRsvps := make(map[Event]int)
	eventRsvps[event] = rsvpCnt

	saveRsvp(inviteCode, phoneNumber, eventRsvps)

	invitedFamily := findInvitedFamily(inviteCode)
	return getFollowupEventAction(invitedFamily, event)
}

func getRsvpCounts(event Event, values map[string]*structpb.Value) int {
	rsvpCnt := -1
	for key, value := range values {
		if CaseInsensitiveContains(key, ".original") {
			continue
		}
		if CaseInsensitiveContains(key, event.Name) && CaseInsensitiveContains(key, "rsvp") {
			rsvpCnt = int(value.GetNumberValue())
			break
		}
	}
	return rsvpCnt
}

func getInviteCodeFromContext(contexts []*dialogflow.Context) int {
	inviteCode := -1
	parameterValue := getFromContext(contexts, "invite_code")
	if parameterValue != nil {
		inviteCode = int(parameterValue.GetNumberValue())
	}
	return inviteCode
}

func getPhoneNumberFromContext(contexts []*dialogflow.Context) string {
	phoneNumber := ""
	parameterValue := getFromContext(contexts, "twilio_sender_id")
	if parameterValue != nil {
		phoneNumber = parameterValue.GetStringValue()
	}
	return phoneNumber
}



func getFromContext(contexts []*dialogflow.Context, givenParameterKey string) *structpb.Value {
	var givenParameterValue *structpb.Value
	for _, c := range contexts {
		for parameterKey, parameterValue := range c.Parameters.GetFields() {
			if parameterKey == givenParameterKey {
				// log.Printf("key: %s value: %s", parameterKey, parameterValue)
				givenParameterValue = parameterValue
				break
			}
		}
	}
	return givenParameterValue
}

func InviteCodeFulfillment(inviteCode int) (string, string) {
	// inviteNumber, err := strconv.Atoi(inviteCode)
	// if err != nil {
	// 	log.Fatalf("Unable to convert invite code(%s) to a number: %v", inviteCode, err)
	// }
	invitedFamily := findInvitedFamily(inviteCode)
	log.Printf("\nReturned Invited_family row %+v", invitedFamily)
	var intent string
	message := fmt.Sprintf("You must be %s.\nYou're invited to", invitedFamily.InviteName)
	if invitedFamily.VidhiInvited > 0 {
		message += eventInviteMsg(Vidhi, invitedFamily.VidhiInvited)
		intent = Vidhi.DialogflowAction
	}
	if invitedFamily.GarbaInvited > 0 {
		message += eventInviteMsg(Garba, invitedFamily.GarbaInvited)
		intent = Garba.DialogflowAction
	}
	if invitedFamily.WeddingInvited > 0 {
		message += eventInviteMsg(Wedding, invitedFamily.WeddingInvited)
		intent = Wedding.DialogflowAction
	}

	message += fmt.Sprintf("\nWould you like to RSVP now?")

	return message, intent
}

func getFollowupEventAction(invitedFamily InvitedFamily, currentEvent Event) string {
	var followupAction string

	switch {
	case invitedFamily.VidhiInvited > 0 && currentEvent != Vidhi:
		followupAction = Vidhi.DialogflowAction
	case invitedFamily.GarbaInvited > 0 && currentEvent != Garba:
		followupAction = Garba.DialogflowAction
	case invitedFamily.WeddingInvited > 0 && currentEvent != Wedding:
		followupAction = Wedding.DialogflowAction
	default:
		log.Printf("Yay!! No more events to rsvp for")
	}

	return followupAction
}

func eventInviteMsg(event Event, invited int) string {
	message := fmt.Sprintf("\n%s: ", event.Name)
	switch invited {
	case NULL_INVITEES:
		message = ""
	case MAX_INVITEES:
		message += "full family"
	default:
		message += strconv.Itoa(invited)
	}
	return message
}

func CaseInsensitiveContains(s, substr string) bool {
	s, substr = strings.ToUpper(s), strings.ToUpper(substr)
	return strings.Contains(s, substr)
}

func findInvitedFamily(inviteNumber int) InvitedFamily {
	wrappedInvitedFamily, _, err := SearchForInvitedFamily(inviteNumber)
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	log.Printf("Invited family for invite code %d - %+v", inviteNumber, wrappedInvitedFamily)

	// Todo: break into separate function
	inviteCodeFromInvitedFamily, _ := convertSheetCellToNumber(wrappedInvitedFamily[3])
	vidhiInvited, _ := convertSheetCellToNumber(wrappedInvitedFamily[4])
	vidhiRsvpd, _ := convertSheetCellToNumber(wrappedInvitedFamily[5])
	garbaInvited, _ := convertSheetCellToNumber(wrappedInvitedFamily[6])
	garbaRsvpd, _ := convertSheetCellToNumber(wrappedInvitedFamily[7])
	weddingInvited, _ := convertSheetCellToNumber(wrappedInvitedFamily[8])
	weddingRsvpd, _ := convertSheetCellToNumber(wrappedInvitedFamily[9])

	return InvitedFamily{
		Origin:         fmt.Sprint(wrappedInvitedFamily[0]),
		Name:           fmt.Sprint(wrappedInvitedFamily[1]),
		InviteName:     fmt.Sprint(wrappedInvitedFamily[2]),
		InviteCode:     inviteCodeFromInvitedFamily,
		VidhiInvited:   vidhiInvited,
		VidhiRsvpd:     vidhiRsvpd,
		GarbaInvited:   garbaInvited,
		GarbaRsvpd:     garbaRsvpd,
		WeddingInvited: weddingInvited,
		WeddingRsvpd:   weddingRsvpd,
	}
}

func convertSheetCellToNumber(data interface{}) (int, error) {
	switch fmt.Sprint(data) {
	case "NULL":
		return 0, nil
	case "ALL":
		return MAX_INVITEES, nil
	default:
		i, err := strconv.Atoi(fmt.Sprint(data))
		return i, err
	}
}

func SearchForInvitedFamily(inviteNumber int) ([]interface{}, int, error) {
	colRange := "A2:J" + strconv.Itoa(TOTAL_INVITED_FAMILY)
	allInvitedFamilies, err := getGoogleSheetsData(INVITED_FAMILY, colRange)
	if err != nil {
		return nil, -1, err
	}

	var invitedFamily []interface{}
	var rowNumber int
	for i, currentInvitedFamily := range allInvitedFamilies {
		// log.Printf("Current invited family: %+v", currentInvitedFamily)
		if len(currentInvitedFamily) > 3 {
			currentInviteNumber, err := convertSheetCellToNumber(currentInvitedFamily[3])
			if err != nil || currentInviteNumber == -1 {
				log.Fatalf("Invite code for entry (%s) wasn't a string or int. Error: +%v", strconv.Itoa(i), err)
			}
			if err == nil && inviteNumber == currentInviteNumber {
				invitedFamily = currentInvitedFamily
				rowNumber = i + 2 // 1 for header & 1 to convert from 0-based to 1-based
				break
			}

		}
	}
	// TODO error not found?

	return invitedFamily, rowNumber, nil
}

func saveRsvp(inviteCode int, phoneNumber string, rsvps map[Event]int) {

	// Save to Update Event
	resp, err := createUpdateEvents(strconv.Itoa(inviteCode), phoneNumber, rsvps)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Http status code for appending an update event: +%v", resp.HTTPStatusCode)

	// Save to Invited Family
	batchResp, batchErr := updateInvitedFamilyRsvp(inviteCode, rsvps)
	if batchErr != nil {
		log.Fatal(batchErr)
	}
	log.Printf("Http status code for updating invited family RSVP: +%v", batchResp.HTTPStatusCode)
}

func createUpdateEvents(inviteCode string, phoneNumber string, rsvps map[Event]int) (*sheets.AppendValuesResponse, error) {
	// Save to Update Event
	var rows [][]interface{}
	for event, attendees := range rsvps {
		var rowData []interface{}
		rowData = append(rowData, inviteCode, phoneNumber, event.Name, attendees, time.Now(), sessionID, responseID)
		rows = append(rows, rowData)
	}

	return appendGoogleSheetsData(UPDATE_EVENT, rows)
}

func updateInvitedFamilyRsvp(inviteCode int, rsvps map[Event]int) (*sheets.BatchUpdateValuesResponse, error) {
	_, rowNumber, err := SearchForInvitedFamily(inviteCode)
	if err != nil {
		log.Fatalf("Unable to update Invited Family rsvp as we can't retrieve the row number: %v", err)
	}

	// Save to Invited Family
	var batchValues []*sheets.ValueRange
	for event, attendees := range rsvps {
		var rowData []interface{}
		var rows [][]interface{}
		rowData = append(rowData, attendees)
		rows = append(rows, rowData)
		writeRange := INVITED_FAMILY + "!" + event.RsvpdCol + strconv.Itoa(rowNumber)
		batchValues = append(batchValues, &sheets.ValueRange{Values: rows, Range: writeRange})
	}
	return setGoogleSheetsData(batchValues)
}

func setGoogleSheetsData(data []*sheets.ValueRange) (*sheets.BatchUpdateValuesResponse, error) {
	rb := &sheets.BatchUpdateValuesRequest{
		ValueInputOption: "USER_ENTERED",
		Data:             data,
	}
	resp, err := getGoogleSheetsClient().Spreadsheets.Values.BatchUpdate(SPREADSHEET_ID, rb).Do()
	return resp, err
}
func appendGoogleSheetsData(sheetName string, rowData [][]interface{}) (*sheets.AppendValuesResponse, error) {
	writeRange := sheetName + "!A2:E2"
	rb := sheets.ValueRange{Values: rowData}
	resp, err := getGoogleSheetsClient().Spreadsheets.Values.Append(SPREADSHEET_ID, writeRange, &rb).ValueInputOption("USER_ENTERED").InsertDataOption("INSERT_ROWS").Do()
	return resp, err
}

func getGoogleSheetsData(sheetName string, colRange string) ([][]interface{}, error) {
	// Retrieve Data
	readRange := sheetName + "!" + colRange
	resp, err := getGoogleSheetsClient().Spreadsheets.Values.Get(SPREADSHEET_ID, readRange).Do()
	// log.Printf("Response from google sheets: +%v", resp)
	return resp.Values, err
}

func getGoogleSheetsClient() *sheets.Service {
	// log.Println("google api creds", os.Getenv("GOOGLE_API_CREDS"))

	// If modifying these scopes, delete your previously saved token.json.
	// Full list of scopes: https://developers.google.com/sheets/api/guides/authorizing
	config, err := google.JWTConfigFromJSON([]byte(os.Getenv("GOOGLE_API_CREDS")), "https://www.googleapis.com/auth/spreadsheets") // Allows read/write access to the user's sheets and their properties.
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := config.Client(oauth2.NoContext)

	srv, err := sheets.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	return srv
}

func main() {
	fmt.Println("Start app")
	// os.Setenv("_LAMBDA_SERVER_PORT", "8080")
	fmt.Println("Start lambda handler")
	lambda.Start(Handler)
}
