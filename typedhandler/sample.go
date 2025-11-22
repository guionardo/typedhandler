package typedhandler

type (
	//easyjson:json
	Request struct {
		Name     string `json:"name"`
		Age      int    `json:"age" `
		City     string `json:"city" query:"city"`
		State    string `json:"state" header:"state"`
		Country  string `json:"country" path:"country"`
		Zip      string `json:"zip"`
		Phone    string `json:"phone" `
		Email    string `json:"email" validate:"email"`
		Password string `json:"password"`
	}
	RequestNormal struct {
		Name     string `json:"name"`
		Age      int    `json:"age" `
		City     string `json:"city" query:"city"`
		State    string `json:"state" header:"state"`
		Country  string `json:"country" path:"country"`
		Zip      string `json:"zip"`
		Phone    string `json:"phone" `
		Email    string `json:"email" validate:"email"`
		Password string `json:"password"`
	}
	Response struct {
		Message string
	}
)

func (r *Request) Reset() {
	r.Name = ""
	r.Age = 0
	r.City = ""
	r.State = ""
}
