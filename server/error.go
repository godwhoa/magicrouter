package server

type HTTPError struct {
	StatusCode int
	Message    string
	Err        error
}

func (e HTTPError) Error() string {
	return e.Err.Error()
}

func (e HTTPError) Unwrap() error {
	return e.Err
}

func (e HTTPError) MarshalJSON() ([]byte, error) {
	return []byte(`{"message": "` + e.Message + `"}`), nil
}
