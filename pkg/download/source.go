package download

type ContentSource struct {
	URL             string `json:"url"`
	Credentials     string `json:"credentials,omitempty"`
	CredentialsType string `json:"credentialsType,omitempty"`
}
