package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	Name       string
	InvitedCol string
	RsvpdCol   string
}

const (
	INVITED_FAMILY       = "INVITED_FAMILY"
	UPDATE_EVENT         = "UPDATE_EVENT"
	TOTAL_INVITED_FAMILY = 240
	SPREADSHEET_ID       = "1FJPePAwh8Xy9revrg8-ANn7GK2Xwd0Xe_6DdLqDujbc"
)

var Vidhi = Event{Name: "VIDHI", InvitedCol: "E", RsvpdCol: "F"}
var Garba = Event{Name: "GARBA", InvitedCol: "G", RsvpdCol: "H"}
var Wedding = Event{Name: "WEDDING", InvitedCol: "I", RsvpdCol: "J"}

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Println("Received body: ", request.Body)

	var err error
	wr := dialogflow.WebhookRequest{}
	unmarshaller := &jsonpb.Unmarshaler{AllowUnknownFields: true}
	if err = unmarshaller.Unmarshal(strings.NewReader(request.Body), &wr); err != nil {
		// if err = jsonpb.UnmarshalString(request.Body, &wr); err != nil {
		// logrus.WithError(err).Error("Couldn't Unmarshal request to jsonpb")
		// c.Status(http.StatusBadRequest)
		// return
		log.Fatal(err)
	}
	// log.Printf("Parsed body: +%v", wr.QueryResult.OutputContexts[0].Parameters)

	switch wr.QueryResult.Intent.DisplayName {
	case "rsvper.rsvp":
		log.Println("Start extracting & saving rsvps")
		rsvps := extractRsvps(wr.QueryResult.OutputContexts[0].Parameters.Fields)
		saveRsvp(241, "8045033244", wr.Session, rsvps)
		break
	}

	return events.APIGatewayProxyResponse{Body: request.Body, StatusCode: 200}, nil
}

func CaseInsensitiveContains(s, substr string) bool {
	s, substr = strings.ToUpper(s), strings.ToUpper(substr)
	return strings.Contains(s, substr)
}

func extractRsvps(values map[string]*structpb.Value) map[Event]int {
	rsvps := make(map[Event]int)
	for key, value := range values {
		if CaseInsensitiveContains(key, ".original") {
			continue
		}

		switch {
		case CaseInsensitiveContains(key, Vidhi.Name):
			rsvps[Vidhi] = int(value.GetNumberValue())
			break
		case CaseInsensitiveContains(key, Garba.Name):
			rsvps[Garba] = int(value.GetNumberValue())
			break
		case CaseInsensitiveContains(key, Wedding.Name):
			rsvps[Wedding] = int(value.GetNumberValue())
			break
		default:
			log.Printf("An unexpected queryResult.outputContexts.parameters key was returned: %s", key)
		}
	}
	return rsvps
}

func findInvitedFamily(inviteCode string) []interface{} {
	inviteNumber, err := strconv.Atoi(inviteCode)
	if err != nil {
		log.Fatalf("Unable to convert invite code(%s) to a number: %v", inviteCode, err)
	}

	// Assume rows are ordered by invite code
	wrappedInvitedFamily, err := getInvitedFamilyRow(inviteNumber + 1) // Add one for the header
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}
	invitedFamily := wrappedInvitedFamily[0]

	// if they are not ordered by invite code
	if invitedFamily == nil || len(invitedFamily) == 0 {
		wrappedInvitedFamily, err := SearchForInvitedFamily(inviteNumber)
		if err != nil {
			log.Fatalf("Unable to retrieve data from sheet: %v", err)
		}
		invitedFamily = wrappedInvitedFamily
	}

	return invitedFamily
}

func SearchForInvitedFamily(inviteNumber int) ([]interface{}, error) {
	colRange := "A2:J" + strconv.Itoa(TOTAL_INVITED_FAMILY)
	allInvitedFamilies, err := getGoogleSheetsData(INVITED_FAMILY, colRange)
	if err != nil {
		return nil, err
	}

	var invitedFamily []interface{}
	for i, currentInvitedFamily := range allInvitedFamilies {
		if len(currentInvitedFamily) > 3 {
			currentInviteNumber, err := extractNumber(currentInvitedFamily[3])
			if err != nil || currentInviteNumber == -1 {
				log.Fatalf("Invite code for entry (%s) wasn't a string or int. Error: +%v", strconv.Itoa(i), err)
			}
			if err == nil && inviteNumber == currentInviteNumber {
				invitedFamily = currentInvitedFamily
				break
			}

		}
	}
	// TODO error not found?

	return invitedFamily, nil
}

func extractNumber(unknownType interface{}) (int, error) {
	var currentInviteNumber int
	switch unknownType.(type) {
	case string:
		wrappedCurrentInviteNumber, err := strconv.Atoi(unknownType.(string))
		if err != nil {
			return -1, err
		}
		currentInviteNumber = wrappedCurrentInviteNumber
		break
	case int:
		currentInviteNumber = unknownType.(int)
		break
	default:
		return -1, errors.New("Unable to convert data to int")
	}
	return currentInviteNumber, nil
}

func getInvitedFamilyRow(rowNumber int) ([][]interface{}, error) {
	colRange := "A" + strconv.Itoa(rowNumber) + ":J" + strconv.Itoa(rowNumber)
	return getGoogleSheetsData(INVITED_FAMILY, colRange) // TODO possible out of range errpr
}

func saveRsvp(inviteCode int, phoneNumber string, sessionID string, rsvps map[Event]int) {

	// Save to Update Event
	resp, err := createUpdateEvents(strconv.Itoa(inviteCode), phoneNumber, sessionID, rsvps)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Http status code for appending an update event: +%v", resp.HTTPStatusCode)

	// Save to Invited Family
	batchResp, batchErr := updateInvitedFamilyRsvp(inviteCode+1, inviteCode, phoneNumber, rsvps)
	if batchErr != nil {
		log.Fatal(batchErr)
	}
	log.Printf("Http status code for updating invited family RSVP: +%v", batchResp.HTTPStatusCode)
}

func createUpdateEvents(inviteCode string, phoneNumber string, sessionID string, rsvps map[Event]int) (*sheets.AppendValuesResponse, error) {
	// Save to Update Event
	var rows [][]interface{}
	for event, attendees := range rsvps {
		var rowData []interface{}
		rowData = append(rowData, inviteCode, phoneNumber, event.Name, attendees, time.Now(), sessionID)
		rows = append(rows, rowData)
	}

	return appendGoogleSheetsData(UPDATE_EVENT, rows)
}

func updateInvitedFamilyRsvp(rowNumber int, inviteCode int, phoneNumber string, rsvps map[Event]int) (*sheets.BatchUpdateValuesResponse, error) {

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

// Handler is our lambda handler invoked by the `lambda.Start` function call
func HandlerOrig(ctx context.Context) (Response, error) {
	var buf bytes.Buffer

	body, err := json.Marshal(map[string]interface{}{
		"message": "Go Serverless v1.0! Your function executed successfully!",
	})
	if err != nil {
		return Response{StatusCode: 404}, err
	}
	json.HTMLEscape(&buf, body)

	resp := Response{
		StatusCode:      200,
		IsBase64Encoded: false,
		Body:            buf.String(),
		Headers: map[string]string{
			"Content-Type":           "application/json",
			"X-MyCompany-Func-Reply": "hello-handler",
		},
	}

	return resp, nil
}

func main() {
	fmt.Println("Start app")
	// os.Setenv("_LAMBDA_SERVER_PORT", "8080")
	fmt.Println("Start lambda handler")
	lambda.Start(Handler)
}
