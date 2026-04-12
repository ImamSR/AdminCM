package models

type StudentDetail struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	NISN      string `json:"nisn"`
	Unit      string `json:"unit"`
	ClassID   int    `json:"class_id"`
	ClassName string `json:"class_name"`
	IsActive  bool   `json:"is_active"`
}

type StudentCreateReq struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	NISN    string `json:"nisn"`
	Unit    string `json:"unit"`
	ClassID int    `json:"class_id"`
}

type StudentUpdateReq struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	NISN  string `json:"nisn"`
	Unit  string `json:"unit"`
}

type StudentStats struct {
	Unit  string `json:"unit"`
	Count int    `json:"count"`
}

type StudentBulkCreateReq struct {
	ClassID  int                `json:"class_id"`
	Unit     string             `json:"unit"`
	Students []StudentCreateReq `json:"students"`
}

type StudentTransferReq struct {
	StudentIDs []int `json:"student_ids"`
	OldClassID int   `json:"old_class_id"`
	NewClassID int   `json:"new_class_id"`
}
