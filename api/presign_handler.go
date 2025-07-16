package api

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
)

type S3PresignHandler struct {
	PresignClient *s3.PresignClient
}

func NewS3PresignHandler(client *s3.PresignClient) *S3PresignHandler {
	return &S3PresignHandler{
		PresignClient: client,
	}
}

type PresignRequest struct {
	FileName string `json:"fileName"`
	FileType string `json:"fileType"`
	FileHash string `json:"fileHash"`
}

type PresignResponse struct {
	URL string `json:"url"`
	Key string `json:"key"`
}

func (p *S3PresignHandler) Upload(w http.ResponseWriter, r *http.Request) {
	var reqBody PresignRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	bucketName := os.Getenv("BUCKET_NAME")
	if bucketName == "" {
		http.Error(w, "Undefined bucket name", http.StatusInternalServerError)
		return
	}

	uuid := uuid.New().String()
	objectKey := "uploads/" + uuid + "_" + reqBody.FileName

	presignRequest, err := p.PresignClient.PresignPutObject(r.Context(), &s3.PutObjectInput{
		Bucket:            aws.String(bucketName),
		Key:               aws.String(objectKey),
		ContentType:       aws.String(reqBody.FileType),
		ChecksumAlgorithm: types.ChecksumAlgorithmSha256,
		ChecksumSHA256:    &reqBody.FileHash,
	}, s3.WithPresignExpires(*aws.Duration(15 * time.Minute)))

	if err != nil {
		log.Printf("Couldn't get a presigned request to put %s in bucket %s. Here's why: %v\n", reqBody.FileName, bucketName, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := PresignResponse{URL: presignRequest.URL, Key: objectKey}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
