package models

type TeacherDetail struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	Unit          string `json:"unit"`
	NIP           string `json:"nip"`
	Qualification string `json:"qualification"`
	Status        string `json:"status"`
	IsActive      bool   `json:"is_active"`
}

type TeacherCreateReq struct {
	Name          string `json:"name"`
	Email         string `json:"email"`
	Unit          string `json:"unit"`
	NIP           string `json:"nip"`
	Qualification string `json:"qualification"`
	Status        string `json:"status"`
}

type TeacherUpdateReq struct {
	Name          string `json:"name"`
	Email         string `json:"email"`
	Unit          string `json:"unit"`
	NIP           string `json:"nip"`
	Qualification string `json:"qualification"`
	Status        string `json:"status"`
}
