package cmd

// maskToken はトークンの一部を隠す
func maskToken(token string) string {
	if token == "" {
		return notSetToken
	}
	if len(token) < minTokenLength {
		return maskedToken
	}
	return token[:4] + "..." + token[len(token)-4:]
}
