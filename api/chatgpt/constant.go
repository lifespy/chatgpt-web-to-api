package chatgpt

//goland:noinspection SpellCheckingInspection
const (
	defaultRole           = "user"
	parseJsonErrorMessage = "Failed to parse json request body."

	csrfUrl        = "https://chat.openai.com/api/auth/csrf"
	promptLoginUrl = "https://chat.openai.com/api/auth/signin/auth0?prompt=login"

	authSessionUrl = "https://chat.openai.com/api/auth/session"

	gpt4Model                          = "gpt-4"
	actionContinue                     = "continue"
	responseTypeMaxTokens              = "max_tokens"
	responseStatusFinishedSuccessfully = "finished_successfully"

	getCsrfTokenErrorMessage     = "获取CSRF token失败."
	getAuthorizedUrlErrorMessage = "获取authorizedUrl失败"
	getStateCodeErrorMessage     = "获取stat code失败"
	getCheckUsernameErrorMessage = "校验邮箱失败"
	getCheckPasswordErrorMessage = "校验密码失败"
	getAccessTokenErrorMessage   = "获取accesstoken失败"
)
