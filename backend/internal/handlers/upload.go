package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

const maxPhotoSize = 5 << 20 // 5 MB

// UploadCandidatePhoto — загрузка фото + AI-детекция
func UploadCandidatePhoto(pool *pgxpool.Pool, uploadDir, geminiAPIKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		candidateID := c.Param("id")

		if err := c.Request.ParseMultipartForm(maxPhotoSize); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file too large (max 5 MB)"})
			return
		}

		file, header, err := c.Request.FormFile("photo")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "photo field is required"})
			return
		}
		defer file.Close()

		contentType := header.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "image/") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "only image files are accepted"})
			return
		}

		fileBytes, err := io.ReadAll(io.LimitReader(file, maxPhotoSize))
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to read file"})
			return
		}

		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			c.JSON(500, gin.H{"error": "failed to create upload directory"})
			return
		}

		ext := filepath.Ext(header.Filename)
		if ext == "" {
			ext = ".jpg"
		}
		filename := fmt.Sprintf("candidate_%s_%d%s", candidateID, time.Now().UnixNano(), ext)
		filePath := filepath.Join(uploadDir, filename)

		if err := os.WriteFile(filePath, fileBytes, 0644); err != nil {
			c.JSON(500, gin.H{"error": "failed to save file"})
			return
		}

		photoURL := "/uploads/" + filename

		aiFlag := false
		aiNote := ""
		if geminiAPIKey != "" {
			aiFlag, aiNote = detectAIGeneratedPhoto(c.Request.Context(), fileBytes, contentType, geminiAPIKey)
		}

		_, err = pool.Exec(c.Request.Context(),
			`UPDATE candidates SET photo_url = $1, photo_ai_flag = $2, photo_ai_note = $3 WHERE id = $4`,
			photoURL, aiFlag, nilIfEmpty(aiNote), candidateID)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to update candidate"})
			return
		}

		c.JSON(200, gin.H{
			"photo_url":    photoURL,
			"ai_flag":      aiFlag,
			"ai_note":      aiNote,
		})
	}
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

const maxDocSize = 10 << 20 // 10 MB

// UploadCandidateDocument — загрузка документа кандидата
func UploadCandidateDocument(pool *pgxpool.Pool, uploadDir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		candidateID := c.Param("id")
		docType := c.Param("docType")

		if err := c.Request.ParseMultipartForm(maxDocSize); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file too large (max 10 MB)"})
			return
		}

		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no file provided"})
			return
		}
		defer file.Close()

		ext := strings.ToLower(filepath.Ext(header.Filename))
		allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".heic": true, ".pdf": true}
		if !allowed[ext] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file type not allowed"})
			return
		}

		filename := fmt.Sprintf("candidate_%s_%s_%d%s", candidateID, docType, time.Now().UnixNano(), ext)
		savePath := filepath.Join(uploadDir, filename)
		out, err := os.Create(savePath)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to save file"})
			return
		}
		defer out.Close()
		if _, err := io.Copy(out, file); err != nil {
			c.JSON(500, gin.H{"error": "failed to save file"})
			return
		}

		fileURL := "/uploads/" + filename

		var column string
		switch docType {
		case "english_cert":
			column = "english_cert_url"
		case "certificate":
			column = "certificate_url"
		case "additional_docs":
			column = "additional_docs_url"
		default:
			c.JSON(400, gin.H{"error": "invalid document type"})
			return
		}

		_, err = pool.Exec(context.Background(),
			fmt.Sprintf("UPDATE candidates SET %s = $1 WHERE id = $2", column),
			fileURL, candidateID)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to update candidate"})
			return
		}

		c.JSON(200, gin.H{"url": fileURL})
	}
}

type geminiVisionRequest struct {
	Contents []geminiVisionContent `json:"contents"`
}

type geminiVisionContent struct {
	Parts []geminiVisionPart `json:"parts"`
}

type geminiVisionPart struct {
	Text       string            `json:"text,omitempty"`
	InlineData *geminiInlineData `json:"inlineData,omitempty"`
}

type geminiInlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type geminiVisionResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func detectAIGeneratedPhoto(ctx context.Context, imgBytes []byte, mimeType, apiKey string) (bool, string) {
	encoded := base64.StdEncoding.EncodeToString(imgBytes)

	prompt := `Analyze this photo and determine if it is AI-generated, digitally manipulated, or contains signs of AI involvement (e.g. GAN artifacts, unnatural skin, blurry backgrounds, wrong facial symmetry, impossible lighting).

Respond ONLY with a JSON object in this exact format:
{"ai_involved": true/false, "confidence": "low|medium|high", "note": "brief explanation in 1 sentence"}

Do not include any other text.`

	reqBody := geminiVisionRequest{
		Contents: []geminiVisionContent{
			{
				Parts: []geminiVisionPart{
					{
						InlineData: &geminiInlineData{
							MimeType: mimeType,
							Data:     encoded,
						},
					},
					{Text: prompt},
				},
			},
		},
	}

	bodyBytes, _ := json.Marshal(reqBody)
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=%s", apiKey)

	httpCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(httpCtx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return false, ""
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, ""
	}
	defer resp.Body.Close()

	var vr geminiVisionResponse
	if err := json.NewDecoder(resp.Body).Decode(&vr); err != nil {
		return false, ""
	}

	if len(vr.Candidates) == 0 || len(vr.Candidates[0].Content.Parts) == 0 {
		return false, ""
	}

	text := strings.TrimSpace(vr.Candidates[0].Content.Parts[0].Text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	var result struct {
		AIInvolved bool   `json:"ai_involved"`
		Confidence string `json:"confidence"`
		Note       string `json:"note"`
	}
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return false, ""
	}

	note := ""
	if result.AIInvolved {
		note = fmt.Sprintf("[%s confidence] %s", result.Confidence, result.Note)
	}
	return result.AIInvolved, note
}
