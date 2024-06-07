package sushi

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/nrf53/makeitippon/internal/pkg/image"
	"golang.org/x/net/websocket"
)

const (
	protoHttps              = "https://"
	protoWebsocket          = "wss://"
	hostSushiski            = "sushi.ski/"
	endpointStream          = "streaming"
	endpointNotesCreate     = "api/notes/create"
	endpointDriveFileCreate = "api/drive/files/create"
	endpointLocal           = "http://localhost:8000"
	outimg                  = "./templates/images/outimage.png"
)

// structs for streaming
type streamRequest struct {
	Type string            `json:"type"`
	Body streamRequestBody `json:"body"`
}

type streamRequestBody struct {
	Channel string `json:"channel"`
	Id      string `json:"id"`
}

// structs for receiving stream response to determine mentions
type streamResponseToCheckType struct {
	Type string                        `json:"type"` // channel
	Body streamResponseBodyToCheckType `json:"body"`
}

type streamResponseBodyToCheckType struct {
	Id   string `json:"id"`   // main
	Type string `json:"type"` // mention
}

// structs for receiving stream response
type streamResponse struct {
	Type string             `json:"type"` // channel
	Body streamResponseBody `json:"body"`
}

type streamResponseBody struct {
	Id   string                 `json:"id"`   // main
	Type string                 `json:"type"` // mention
	Body streamResponseBodyBody `json:body`
}
type streamResponseBodyBody struct {
	Id        string              `json:"id"`
	CreatedAt string              `json:"createdAt"`
	UserId    string              `json:"userId"`
	User      streamResponseUser  `json:user`
	Text      string              `json:"text`
	Mentions  []string            `json:mentions`
	Reply     streamResponseReply `json:reply`
}
type streamResponseReply struct {
	Id        string             `json:"id"`
	CreatedAt string             `json:"createdAt"`
	UserId    string             `json:"userId"`
	User      streamResponseUser `json:user`
	Text      string             `json:"text`
}

type streamResponseUser struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
}

// struct for note date
type noteData struct {
	MentionNoteId    string
	mentionReplyText string
}

type driveResponse struct {
	Id        string `json:"id"`
	CreatedAt string `json:"createdAt"`
	Name      string `json:"name"`
	Size      int    `json:"size"`
}

// structs for replying and files
type postReplyRequestBody struct {
	Visibility string `json:"visibility"`
	ReplyID    string `json:"replyId"`
	//Text       string   `json:"text"`
	FileIds []string `json:"fileIds"`
}

func getAccessTokenSushiski() string {
	// Read access token from ENV
	accessTokenSushiski := os.Getenv("ACCESS_TOKEN_SUSHISKI")

	return accessTokenSushiski
}

// dial to streaming websocket
func dialWebsocket() *websocket.Conn {

	// get ACCESS_TOKEN_SUSHISKI
	accessTokenSushiski := getAccessTokenSushiski()

	// create websocket connection
	var err error
	ws, _ := websocket.Dial(protoWebsocket+hostSushiski+endpointStream+"?i="+accessTokenSushiski, "", endpointLocal)

	if err != nil {
		log.Fatal(err)
	}
	// close ws
	//defer ws1.Close()
	log.Println("[INFO] Completed to dial to Websocket connection.")

	// prepare message for websocket connection
	var streamRequestMain streamRequest
	streamRequestMain.Type = "connect"
	streamRequestMain.Body.Channel = "main"
	streamRequestMain.Body.Id = "main"

	streamRequestMainBytes, _ := json.Marshal(streamRequestMain)

	var streamRequestHome streamRequest
	streamRequestHome.Type = "connect"
	streamRequestHome.Body.Channel = "homeTimeline"
	streamRequestHome.Body.Id = "homeTimeline"

	streamRequestHomeBytes, _ := json.Marshal(streamRequestHome)

	// send message to websocket connection
	websocket.Message.Send(ws, streamRequestMainBytes)
	websocket.Message.Send(ws, streamRequestHomeBytes)
	// return websocket connection pointer

	/* for debug
	log.Println("dialWebsocket()")
	fmt.Println(reflect.TypeOf(ws))
	fmt.Printf("%v\n", ws)
	*/

	log.Println("[INFO] Connected to streaming.")
	return ws
}

// receiving message from websocket
func receiveMessage(ws *websocket.Conn, streamResponseMsg *string) {
	/* for debug
	log.Println("receiveMessage()")
	fmt.Println(reflect.TypeOf(ws))
	fmt.Printf("%v\n", ws)
	println(ws.RemoteAddr().String())
	*/
	err := websocket.Message.Receive(ws, streamResponseMsg)
	if err != nil {
		log.Println("[ERROR] Failed to receive message from Websocket")
		log.Fatal(err)
	}
}

// checking if the message is a mention
func isMessageMention(streamResponseMsg string) bool {
	// prepare for streamResponseToCheckType
	var streamResponseBodyToCheckType streamResponseToCheckType

	// unmarshal json message to streamResponseToCheckType
	err := json.Unmarshal([]byte(streamResponseMsg), &streamResponseBodyToCheckType)
	if err != nil {
		log.Println("[ERROR] Failed to unmarshal json message to struct when check if the message is a mention")
		log.Fatal(err)
	}

	// check type
	if streamResponseBodyToCheckType.Body.Type == "mention" {
		/* for debug
		log.Println("isMessageMention()")
		fmt.Println("true")
		*/
		return true
	} else {
		/* for debug
		log.Println("isMessageMention()")
		fmt.Println("false")
		*/
		return false
	}
}

// get note date
func getNoteData(streamResponseMsg string, noteData *noteData) {
	// prepare struct
	var streamRespons streamResponse

	// unmarshal json message to struct
	err := json.Unmarshal([]byte(streamResponseMsg), &streamRespons)
	if err != nil {
		log.Println("[ERROR] Failed to unmarshal json message to struct")
		log.Fatal(err)
	}

	// store necessary date
	noteData.MentionNoteId = streamRespons.Body.Body.Id            // Id for bot to reply to
	noteData.mentionReplyText = streamRespons.Body.Body.Reply.Text // original text to be transformed to png

	/* for debug
	log.Println("getNoteData()")
	fmt.Println(reflect.TypeOf(streamRespons.Body.Body.Reply.Text))
	fmt.Printf("%v\n", streamRespons.Body.Body.Reply.Text)
	*/

}

// upload image to drive
func uploadImageToDrive() string {
	// prepare file

	file, err := os.Open(outimg)
	if err != nil {
		log.Println("[ERROR] Failed to open image")
		log.Fatal(err)
	}
	defer file.Close()

	// construct multi data form
	requestBody := &bytes.Buffer{}
	multiDataFormWriter := multipart.NewWriter(requestBody)
	part, _ := multiDataFormWriter.CreateFormFile("file", outimg)
	io.Copy(part, file)
	multiDataFormWriter.Close()

	// prepare http request
	uploadRequest, err := http.NewRequest(
		"POST",
		protoHttps+hostSushiski+endpointDriveFileCreate,
		requestBody,
	)
	if err != nil {
		log.Println("[ERROR] Failed to prepare http request")
		log.Fatal(err)
	}

	// set http headers
	uploadRequest.Header.Set("Authorization", "Bearer "+getAccessTokenSushiski())
	uploadRequest.Header.Set("Content-type", multiDataFormWriter.FormDataContentType())

	// http request
	client := &http.Client{}
	uploadResponse, err := client.Do(uploadRequest)
	if err != nil {
		log.Println("[ERROR] Failed to complete http request")
		log.Fatal(err)
	}
	defer uploadResponse.Body.Close()

	// parse http response
	var driveRespnseBody driveResponse
	jsonBytes, err := io.ReadAll(uploadResponse.Body)
	err = json.Unmarshal(jsonBytes, &driveRespnseBody)
	if err != nil {
		log.Println("[ERROR] Failed to parse response json")
		log.Fatal(err)
	}

	// return file id
	return driveRespnseBody.Id
}

// post reply note with image
func postReplyNote(fileId string, noteData *noteData) {
	// prepare struct
	var postReplyRequestBody postReplyRequestBody

	// construct request body
	postReplyRequestBody.Visibility = "followers"
	postReplyRequestBody.ReplyID = noteData.MentionNoteId
	postReplyRequestBody.FileIds = append(postReplyRequestBody.FileIds, fileId)

	postReplyRequestBodyBytes, err := json.Marshal(postReplyRequestBody)
	if err != nil {
		log.Println("[ERROR] Failed marshal struct to json")
		log.Fatal(err)
	}

	// prepare http request
	replyRequest, err := http.NewRequest(
		"POST",
		protoHttps+hostSushiski+endpointNotesCreate,
		bytes.NewBuffer(postReplyRequestBodyBytes),
	)
	if err != nil {
		log.Println("[ERROR] Failed to prepare http request")
		log.Fatal(err)
	}

	// set http headers
	replyRequest.Header.Set("Authorization", "Bearer "+getAccessTokenSushiski())
	replyRequest.Header.Set("Content-type", "application/json")

	// http request
	client := &http.Client{}
	replyResponse, err := client.Do(replyRequest)
	if err != nil {
		log.Println("[ERROR] Failed to complete http request")
		log.Fatal(err)
	}
	defer replyResponse.Body.Close()

}

func Main() {
	// prepare websocket connection

	// dial websocket connection
	ws := dialWebsocket()
	log.Println("[INFO] Completed to establish Websocket connection.")

	/* for debug
	log.Println("Main()")
	fmt.Println(reflect.TypeOf(ws))
	fmt.Printf("%v\n", ws)
	*/

	// stream websocket connection
	for {
		// receiving message
		var streamResponseMsg string
		receiveMessage(ws, &streamResponseMsg)

		// check if the message is a mention
		if isMessageMention(streamResponseMsg) {
			log.Println("[INFO] Received mention.")

			// prepare NoteData struct
			var noteData noteData
			getNoteData(streamResponseMsg, &noteData)

			// generate image
			image.Text2img(noteData.mentionReplyText)
			log.Println("[INFO] Generated image.")

			// upload image to drive
			fileId := uploadImageToDrive()
			log.Println("[INFO] Uploaded image.")

			// post reply note with image
			postReplyNote(fileId, &noteData)
			log.Println("[INFO] Posted MAKEITIPPON.")

		}
	}
	// defer ws.Close()
}
