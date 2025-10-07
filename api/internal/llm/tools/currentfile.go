package tools

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
)

type CurrentFileParams struct {
	FilePath string `json:"file_path"`
}

type currentFileTool struct {
	app        CurrentFileToolbox
	workingDir string
}

const CurrentFileToolName = "currentfile"

//go:embed currentfile.md
var currentFileDescription []byte

func NewCurrentFileTool(app CurrentFileToolbox, workingDir string) BaseTool {
	return &currentFileTool{
		app:        app,
		workingDir: workingDir,
	}
}

// CurrentFileToolbox define a interface que a ferramenta precisa para interagir com o estado da aplicação.
type CurrentFileToolbox interface {
	SetCurrentFile(file string)
}

func (t *currentFileTool) Name() string {
	return CurrentFileToolName
}

func (t *currentFileTool) Info() ToolInfo {
	return ToolInfo{
		Name:        CurrentFileToolName,
		Description: string(currentFileDescription),
		Parameters: map[string]any{
			"file_path": map[string]any{
				"type":        "string",
				"description": "O caminho do arquivo a ser definido como o arquivo atual.",
			},
		},
		Required: []string{"file_path"},
	}
}

func (t *currentFileTool) Run(ctx context.Context, call ToolCall) (ToolResponse, error) {
	var params CurrentFileParams
	if err := json.Unmarshal([]byte(call.Input), &params); err != nil {
		return NewTextErrorResponse(fmt.Sprintf("error parsing parameters: %s", err)), nil
	}

	t.app.SetCurrentFile(params.FilePath)
	return NewTextResponse(fmt.Sprintf("Arquivo atual definido como '%s'.", params.FilePath)), nil
}
