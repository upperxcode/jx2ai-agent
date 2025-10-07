package api

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/johnxcode/jx2ai-agent/internal/llm/tools"
)

// FileInfo representa a informação de um arquivo ou diretório para o frontend.
type FileInfo struct {
	Name  string `json:"name"`
	IsDir bool   `json:"isDir"`
}

// CommandInfo descreve um comando para o frontend.
type CommandInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UIState representa o estado atual da interface que o backend gerencia.
type UIState struct {
	CurrentFile   string   `json:"currentFile"`
	AttachedFiles []string `json:"attachedFiles"`
}

// App struct
type App struct {
	ctx           context.Context
	currentFile   string
	attachedFiles map[string]bool // Usando um map para facilitar a adição/remoção e evitar duplicatas

	toolbelt map[string]tools.BaseTool
}

// NewApp creates a new App application struct
func NewApp() (*App, error) {
	app := &App{
		attachedFiles: make(map[string]bool),
	}

	// Instancia as ferramentas, passando a própria app como dependência.
	// A `app` implementa as interfaces que as ferramentas precisam.
	app.toolbelt = map[string]tools.BaseTool{
		tools.AttachToolName:      tools.NewAttachTool(app, ""), // workingDir não é necessário aqui
		tools.CurrentFileToolName: tools.NewCurrentFileTool(app, ""),
	}

	return app, nil
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// ListCommands retorna a lista de comandos disponíveis para o frontend.
// Isso simula a leitura dos arquivos de descrição das ferramentas.
func (a *App) ListCommands() []CommandInfo {
	return []CommandInfo{
		{Name: "currentfile", Description: "Define o arquivo principal para interação com a IA."},
		{Name: "attach", Description: "Anexa ou remove arquivos do contexto da conversa."},
	}
}

// ExecuteCommand recebe uma string de comando do frontend, a processa e retorna o novo estado da UI.
func (a *App) ExecuteCommand(commandString string) (UIState, error) {
	parts := strings.Fields(strings.TrimPrefix(commandString, "/"))
	if len(parts) == 0 {
		return UIState{}, fmt.Errorf("comando inválido")
	}

	command := parts[0]
	args := parts[1:]

	tool, ok := a.toolbelt[command]
	if !ok {
		return UIState{}, fmt.Errorf("comando desconhecido: %s", command)
	}

	// Constrói o input para a ferramenta
	var input string
	if command == "attach" && len(args) >= 2 {
		input = fmt.Sprintf(`{"sub_command": "%s", "file_path": "%s"}`, args[0], args[1])
	} else if command == "currentfile" && len(args) >= 1 {
		input = fmt.Sprintf(`{"file_path": "%s"}`, args[0])
	} else {
		return UIState{}, fmt.Errorf("argumentos inválidos para o comando %s", command)
	}

	_, err := tool.Run(a.ctx, tools.ToolCall{Input: input})
	if err != nil {
		return UIState{}, fmt.Errorf("erro ao executar o comando %s: %w", command, err)
	}

	return a.getCurrentUIState(), nil
}

// getCurrentUIState coleta o estado atual e o prepara para ser enviado ao frontend.
func (a *App) getCurrentUIState() UIState {
	// Converte o map de arquivos anexados em uma lista de strings
	attached := make([]string, 0, len(a.attachedFiles))
	for file := range a.attachedFiles {
		attached = append(attached, file)
	}
	// Ordena para uma exibição consistente na UI
	sort.Strings(attached)

	var currentFileDisplay string
	if a.currentFile != "" {
		currentFileDisplay = a.currentFile
	}

	return UIState{
		CurrentFile:   currentFileDisplay,
		AttachedFiles: attached,
	}
}

// SetCurrentFile define o arquivo atual.
func (a *App) SetCurrentFile(file string) {
	a.currentFile = file
}

// AttachFile adiciona um arquivo à lista de anexos.
func (a *App) AttachFile(file string) {
	a.attachedFiles[file] = true
}

// DetachFile remove um arquivo da lista de anexos.
func (a *App) DetachFile(file string) {
	delete(a.attachedFiles, file)
}

// ListDirectory retorna uma lista de arquivos e diretórios para um dado caminho.
// A lista é ordenada alfabeticamente, com arquivos primeiro e depois diretórios.
func (a *App) ListDirectory(path string) ([]FileInfo, error) {
	// Medida de segurança: normaliza o caminho para evitar ataques de "path traversal"
	cleanPath := filepath.Clean(path)

	// Obtém o diretório de trabalho atual como base
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter o diretório de trabalho: %w", err)
	}

	// Constrói o caminho completo e seguro
	fullPath := filepath.Join(cwd, cleanPath)

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler o diretório '%s': %w", fullPath, err)
	}

	var files []FileInfo
	var dirs []FileInfo

	for _, entry := range entries {
		info := FileInfo{
			Name:  entry.Name(),
			IsDir: entry.IsDir(),
		}
		if entry.IsDir() {
			dirs = append(dirs, info)
		} else {
			files = append(files, info)
		}
	}

	// A lista final terá os arquivos ordenados e depois os diretórios ordenados
	return append(files, dirs...), nil
}

// ReadFile lê o conteúdo de um arquivo e o retorna como uma string.
func (a *App) ReadFile(path string) (string, error) {
	// Medida de segurança: normaliza o caminho para evitar ataques de "path traversal"
	cleanPath := filepath.Clean(path)

	// Obtém o diretório de trabalho atual como base
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("erro ao obter o diretório de trabalho: %w", err)
	}

	// Constrói o caminho completo e seguro
	fullPath := filepath.Join(cwd, cleanPath)

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("erro ao ler o arquivo '%s': %w", fullPath, err)
	}

	return string(content), nil
}
