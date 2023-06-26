package chatgpt

import (
	"bufio"
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/linweiyuan/go-chatgpt-api/api"
	uuid "github.com/satori/go.uuid"
	"strings"

	http "github.com/bogdanfinn/fhttp"
)

var (
	arkoseTokenUrl string
)

//goland:noinspection SpellCheckingInspection
func init() {
	arkoseTokenUrl = "http://arkosetoken.api.xiu.ee/"
}

type simpleConversationRequest struct {
	Message string `json:"message"`
	Model   string `json:"model"`
}

//goland:noinspection GoUnhandledErrorResult
func CreateConversationSimple(c *gin.Context) {
	var request simpleConversationRequest
	if err := c.BindJSON(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, api.ReturnMessage(parseJsonErrorMessage))
		return
	}

	param := CreateConversationRequest{
		Action: "next",
		Messages: []Message{
			{
				ID:     uuid.NewV4().String(),
				Author: Author{Role: defaultRole},
				Content: Content{
					ContentType: "text",
					Parts:       []string{request.Message},
				},
			},
		},
		Model:                      request.Model,
		TimezoneOffsetMin:          -480,
		HistoryAndTrainingDisabled: false,
	}

	if param.Model == "" {
		param.Model = gpt4Model
	}

	if param.ConversationID == nil || *param.ConversationID == "" {
		param.ConversationID = nil
	}

	if len(param.Messages) != 0 {
		if param.Messages[0].Author.Role == "" {
			param.Messages[0].Author.Role = defaultRole
		}
	}

	if strings.HasPrefix(param.Model, gpt4Model) {
		if arkoseTokenUrl != "" {
			req, _ := http.NewRequest(http.MethodGet, arkoseTokenUrl, nil)
			resp, err := api.Client.Do(req)
			if err != nil || resp.StatusCode != http.StatusOK {
				c.AbortWithStatusJSON(http.StatusInternalServerError, api.ReturnMessage("Failed to get arkose token."))
				return
			}

			responseMap := make(map[string]string)
			json.NewDecoder(resp.Body).Decode(&responseMap)
			param.ArkoseToken = responseMap["token"]
		}
	}

	resp, done := sendConversationRequest(c, param)
	if done {
		return
	}

	handleConversationResponse(c, resp, param)
}

//goland:noinspection GoUnhandledErrorResult
func CreateConversation(c *gin.Context) {
	var request CreateConversationRequest
	if err := c.BindJSON(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, api.ReturnMessage(parseJsonErrorMessage))
		return
	}

	if request.ConversationID == nil || *request.ConversationID == "" {
		request.ConversationID = nil
	}

	if len(request.Messages) != 0 {
		if request.Messages[0].Author.Role == "" {
			request.Messages[0].Author.Role = defaultRole
		}
	}

	if strings.HasPrefix(request.Model, gpt4Model) {
		if arkoseTokenUrl != "" {
			req, _ := http.NewRequest(http.MethodGet, arkoseTokenUrl, nil)
			resp, err := api.Client.Do(req)
			if err != nil || resp.StatusCode != http.StatusOK {
				c.AbortWithStatusJSON(http.StatusInternalServerError, api.ReturnMessage("Failed to get arkose token."))
				return
			}

			responseMap := make(map[string]string)
			json.NewDecoder(resp.Body).Decode(&responseMap)
			request.ArkoseToken = responseMap["token"]
		}
	}

	resp, done := sendConversationRequest(c, request)
	if done {
		return
	}

	handleConversationResponse(c, resp, request)
}

//goland:noinspection GoUnhandledErrorResult
func sendConversationRequest(c *gin.Context, request CreateConversationRequest) (*http.Response, bool) {
	jsonBytes, _ := json.Marshal(request)
	req, _ := http.NewRequest(http.MethodPost, api.ChatGPTApiUrlPrefix+"/backend-api/conversation", bytes.NewBuffer(jsonBytes))
	req.Header.Set("User-Agent", api.UserAgent)
	req.Header.Set("Accept", "text/event-stream")
	accessToken := TokenManager.GetToken()
	req.Header.Set("Authorization", api.GetAccessToken(accessToken.AccessToken))
	// Clear cookies
	//if accessToken.PUID != "" {
	//	req.Header.Set("Cookie", "_puid="+accessToken.PUID+";")
	//}
	resp, err := api.Client.Do(req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, api.ReturnMessage(err.Error()))
		return nil, true
	}

	if resp.StatusCode != http.StatusOK {
		responseMap := make(map[string]interface{})
		json.NewDecoder(resp.Body).Decode(&responseMap)
		c.AbortWithStatusJSON(resp.StatusCode, responseMap)
		return nil, true
	}

	return resp, false
}

//goland:noinspection GoUnhandledErrorResult
func handleConversationResponse(c *gin.Context, resp *http.Response, request CreateConversationRequest) {
	c.Writer.Header().Set("Content-Type", "text/event-stream; charset=utf-8")

	isMaxTokens := false
	continueParentMessageID := ""
	continueConversationID := ""

	defer resp.Body.Close()
	reader := bufio.NewReader(resp.Body)
	for {
		if c.Request.Context().Err() != nil {
			break
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "event") ||
			strings.HasPrefix(line, "data: 20") ||
			line == "" {
			continue
		}

		responseJson := line[6:]
		if strings.HasPrefix(responseJson, "[DONE]") && isMaxTokens && request.AutoContinue {
			continue
		}

		// no need to unmarshal every time, but if response content has this "max_tokens", need to further check
		if strings.TrimSpace(responseJson) != "" && strings.Contains(responseJson, responseTypeMaxTokens) {
			var createConversationResponse CreateConversationResponse
			json.Unmarshal([]byte(responseJson), &createConversationResponse)
			message := createConversationResponse.Message
			if message.Metadata.FinishDetails.Type == responseTypeMaxTokens && createConversationResponse.Message.Status == responseStatusFinishedSuccessfully {
				isMaxTokens = true
				continueParentMessageID = message.ID
				continueConversationID = createConversationResponse.ConversationID
			}
		}

		c.Writer.Write([]byte(line + "\n\n"))
		c.Writer.Flush()
	}

	if isMaxTokens && request.AutoContinue {
		continueConversationRequest := CreateConversationRequest{
			ArkoseToken:                request.ArkoseToken,
			HistoryAndTrainingDisabled: request.HistoryAndTrainingDisabled,
			Model:                      request.Model,
			TimezoneOffsetMin:          request.TimezoneOffsetMin,

			Action:          actionContinue,
			ParentMessageID: continueParentMessageID,
			ConversationID:  &continueConversationID,
		}
		resp, done := sendConversationRequest(c, continueConversationRequest)
		if done {
			return
		}

		handleConversationResponse(c, resp, continueConversationRequest)
	}
}
