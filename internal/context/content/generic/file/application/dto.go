package application

// UploadFileRequest is the DTO for file upload.
type UploadFileRequest struct {
	OriginalName string `json:"original_name"`
	MimeType     string `json:"mime_type"`
	Size         int64  `json:"size"`
	UploadedBy   string `json:"uploaded_by"`
}
