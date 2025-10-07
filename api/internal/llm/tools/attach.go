package tools

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
)

type AttachParams struct {
	SubCommand string `json:"sub_command"` // "add" or "delete"
	FilePath   string `json:"file_path"`
}

type attachTool struct {
	app        Toolbox
	workingDir string
}

const AttachToolName = "attach"

//go:embed attach.md
var attachDescription []byte

func NewAttachTool(app Toolbox, workingDir string) BaseTool {
	return &attachTool{
		app:        app,
		workingDir: workingDir,
	}
}

// Toolbox define a interface que a ferramenta precisa para interagir com o estado da aplicação.
type Toolbox interface {
	AttachFile(file string)
	DetachFile(file string)
}

func (t *attachTool) Name() string {
	return AttachToolName
}

func (t *attachTool) Info() ToolInfo {
	return ToolInfo{
		Name:        AttachToolName,
		Description: string(attachDescription),
		Parameters: map[string]any{
			"sub_command": map[string]any{
				"type":        "string",
				"description": "Subcomando 'add' para anexar ou 'delete' para remover.",
			},
			"file_path": map[string]any{
				"type":        "string",
				"description": "O caminho do arquivo para anexar ou remover.",
			},
		},
		Required: []string{"sub_command", "file_path"},
	}
}

func (t *attachTool) Run(ctx context.Context, call ToolCall) (ToolResponse, error) {
	var params AttachParams
	if err := json.Unmarshal([]byte(call.Input), &params); err != nil {
		return NewTextErrorResponse(fmt.Sprintf("error parsing parameters: %s", err)), nil
	}

	switch params.SubCommand {
	case "add":
		t.app.AttachFile(params.FilePath)
		return NewTextResponse(fmt.Sprintf("Arquivo '%s' anexado.", params.FilePath)), nil
	case "delete":
		t.app.DetachFile(params.FilePath)
		return NewTextResponse(fmt.Sprintf("Arquivo '%s' desanexado.", params.FilePath)), nil
	default:
		return NewTextErrorResponse(fmt.Sprintf("Subcomando desconhecido: %s. Use 'add' ou 'delete'.", params.SubCommand)), nil
	}
}
