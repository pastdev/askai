package image

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pastdev/askai/pkg/log"
	"github.com/pastdev/open/pkg/open"
	"github.com/sashabaranov/go-openai"
)

type ResponseWriter interface {
	Write(openai.ImageResponse) error
}

type FileResponseWriter struct {
	Dir        string
	HTTPClient openai.HTTPDoer
	Open       bool
}

type RawResponseWriter struct {
	Open bool
	W    io.Writer
}

func (w *FileResponseWriter) Write(res openai.ImageResponse) error {
	for i, data := range res.Data {
		switch {
		case data.B64JSON != "":
			return w.writeB64JSON(i, data.B64JSON)
		case data.URL != "":
			return w.writeURL(i, data.URL)
		default:
			return fmt.Errorf("unknown image data format")
		}
	}
	return nil
}

func (w *FileResponseWriter) writeB64JSON(index int, b64JSON string) error {
	data, err := base64.StdEncoding.DecodeString(b64JSON)
	if err != nil {
		return fmt.Errorf("writeb64json decode: %w", err)
	}

	return w.writeData(index, data)
}

func (w *FileResponseWriter) writeData(index int, data []byte) error {
	var ext string
	contentType := http.DetectContentType(data)
	switch contentType {
	case "image/gif":
		ext = "gif"
	case "image/jpeg":
		ext = "jpg"
	case "image/png":
		ext = "png"
	default:
		return fmt.Errorf("writeb64json unsupported content type: %s", contentType)
	}

	file := filepath.Join(w.Dir, fmt.Sprintf("%d.%s", index, ext))
	err := os.MkdirAll(w.Dir, 0700)
	if err != nil {
		return fmt.Errorf("writeb64json mkdir: %w", err)
	}

	log.Debug().Str("file", file).Msg("write b64json image data")
	err = os.WriteFile(file, data, 0600)
	if err != nil {
		return fmt.Errorf("writeb64json write: %w", err)
	}

	if w.Open {
		_, err := open.Open(file)
		if err != nil {
			return fmt.Errorf("writeb64json open: %w", err)
		}
	}

	return nil
}

func (w *FileResponseWriter) writeURL(index int, imageURL string) error {
	req, err := http.NewRequest("GET", imageURL, nil)
	if err != nil {
		return fmt.Errorf("writeurl new request: %w", err)
	}

	resp, err := w.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("writeurl download: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("writeurl download: %w", err)
	}

	return w.writeData(index, data)
}

// Writes the response as JSON.
// Example response:
//
//	{"data":[{"url":"https://imgen.x.ai/xai-imgen/xai-tmp-imgen-ac6e2ba9-5acf-4809-85f2-f408f5747288.jpeg","revised_prompt":"A high-resolution photograph of a black Labrador Retriever with white spots, taken in a lush green park during the day. The dog is the central focus, sitting calmly on grass with a slightly alert expression, facing forward. The background features trees and a clear sky, providing a natural and serene setting without distractions. The lighting is soft and natural, highlighting the dog's fur texture and color. The composition avoids any foreground elements that might distract from the main subject, ensuring the dog's unique appearance is the highlight of the image."}]}
func (w *RawResponseWriter) Write(res openai.ImageResponse) error {
	err := json.NewEncoder(w.W).Encode(res)
	if err != nil {
		return fmt.Errorf("write encode: %w", err)
	}

	if w.Open {
		for _, data := range res.Data {
			switch {
			case data.B64JSON != "":
				return fmt.Errorf("write open b64 not supported in raw response writer: %w", err)
			case data.URL != "":
				_, err := open.Open(data.URL)
				if err != nil {
					return fmt.Errorf("write open url: %w", err)
				}
			default:
				return fmt.Errorf("unknown image data format")
			}
		}
	}

	return nil
}
