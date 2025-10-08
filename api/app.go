package api

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/upperxcode/jx2ai-agent/api/internal/config"
	"github.com/upperxcode/jx2ai-agent/api/internal/csync"
	"github.com/upperxcode/jx2ai-agent/api/internal/db"
	"github.com/upperxcode/jx2ai-agent/api/internal/env"
	"github.com/upperxcode/jx2ai-agent/api/internal/history"
	"github.com/upperxcode/jx2ai-agent/api/internal/llm/tools"
	"github.com/upperxcode/jx2ai-agent/api/internal/lsp"
	"github.com/upperxcode/jx2ai-agent/api/internal/permission"
)

// FileInfo representa a informação de um arquivo ou diretório para o frontend.
type FileInfo struct {
	Name  string `json:"name"`
	IsDir bool   `json:"isDir"`
}

func Config() (*config.Config, error) {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Erro ao obter o diretório de trabalho: %v", err)
		return nil, err
	}
	c, err := config.Init(wd, "", false)
	if err != nil {
		panic("config not loaded")
	}
	return c, nil
}

// CommandInfo descreve um comando para o frontend.
type CommandInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UIState representa o estado atual da interface que o backend gerencia.
type UIState struct {
	CurrentFile   string   `json:"currentFile"`   // O arquivo atualmente em foco.
	AttachedFiles []string `json:"attachedFiles"` // Lista de arquivos anexados para contexto.
	ViewContent   string   `json:"viewContent"`   // Conteúdo para ser exibido diretamente (usado pelo comando /view).
}

// App struct
type App struct {
	ctx           context.Context
	currentFile   string
	attachedFiles map[string]bool // Usando um map para facilitar a adição/remoção e evitar duplicatas
	toolbelt      map[string]tools.BaseTool
	config        *config.Config
	lspClients    *csync.Map[string, *lsp.Client]
	permissions   permission.Service
	files         history.Service
}

// NewApp creates a new App application struct
func NewApp() (*App, error) {
	cfg := config.Get()
	if cfg == nil {
		return nil, fmt.Errorf("config not loaded")
	}

	app := &App{
		config:        cfg,
		attachedFiles: make(map[string]bool),
		lspClients:    csync.NewMap[string, *lsp.Client](),
	}

	// Instancia as ferramentas, passando a própria app como dependência.
	// A `app` implementa as interfaces que as ferramentas precisam.
	app.toolbelt = map[string]tools.BaseTool{
		tools.AttachToolName:      tools.NewAttachTool(app, ""), // workingDir não é necessário aqui
		tools.CurrentFileToolName: tools.NewCurrentFileTool(app, ""),
		tools.ViewToolName:        tools.NewViewTool(nil, nil, ""),
		tools.WriteToolName:       tools.NewWriteTool(nil, nil, nil, ""),
	}

	return app, nil
}

// NewAppWithServices cria uma nova App com todos os serviços necessários inicializados.
func NewAppWithServices(ctx context.Context) (*App, error) {
	cfg := config.Get()
	if cfg == nil {
		return nil, fmt.Errorf("configuração não carregada")
	}

	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter o diretório de trabalho: %w", err)
	}
	allowed := []string{tools.ViewToolName, tools.CurrentFileToolName}

	conn, err := db.Connect(ctx, cfg.Options.DataDirectory)
	if err != nil {
		return nil, err
	}

	q := db.New(conn)

	app := &App{
		ctx:           ctx,
		config:        cfg,
		attachedFiles: make(map[string]bool),
		lspClients:    csync.NewMap[string, *lsp.Client](),
		permissions:   permission.NewPermissionService(wd, true, allowed),
		files:         history.NewService(q, conn),
	}

	// Inicializa os clientes LSP
	for name, lspConfig := range cfg.LSP {
		if !lsp.HasRootMarkers(wd, lspConfig.RootMarkers) {
			slog.Debug("LSP server not configured for this project", "server", name)
			continue
		}
		client, err := lsp.New(ctx, name, lspConfig, config.NewEnvironmentVariableResolver(env.New()))
		if err != nil {
			if exec.IsNotFound(err) {
				slog.Warn("LSP server command not found", "server", name, "command", lspConfig.Command)
				continue
			}
			return nil, fmt.Errorf("falha ao criar o cliente LSP %s: %w", name, err)
		}
		app.lspClients.Set(name, client)
	}

	app.toolbelt = map[string]tools.BaseTool{
		tools.AttachToolName:      tools.NewAttachTool(app, wd),
		tools.CurrentFileToolName: tools.NewCurrentFileTool(app, wd),
		tools.ViewToolName:        tools.NewViewTool(app.lspClients, app.permissions, wd),
		tools.WriteToolName:       tools.NewWriteTool(app.lspClients, app.permissions, app.files, wd),
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
	var commands []CommandInfo
	for _, tool := range a.toolbelt {
		info := tool.Info()
		commands = append(commands, CommandInfo{Name: info.Name, Description: info.Description})
	}
	return commands
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
	switch command {
	case tools.AttachToolName:
		if len(args) >= 2 {
			input = fmt.Sprintf(`{"sub_command": %s, "file_path": %s}`, strconv.Quote(args[0]), strconv.Quote(args[1]))
		} else {
			return UIState{}, fmt.Errorf("o comando 'attach' requer um subcomando (add/delete) e um caminho de arquivo")
		}
	case tools.CurrentFileToolName:
		if len(args) >= 1 {
			input = fmt.Sprintf(`{"file_path": %s}`, strconv.Quote(args[0]))
		} else {
			return UIState{}, fmt.Errorf("o comando 'currentfile' requer um caminho de arquivo")
		}
	case tools.WriteToolName:
		if len(args) >= 1 {
			input = fmt.Sprintf(`{"file_path": %s, "content": %s}`, strconv.Quote(args[0]), strconv.Quote(strings.Join(args[1:], " ")))
		} else {
			return UIState{}, fmt.Errorf("o comando 'write' requer um caminho de arquivo")
		}
	case tools.ViewToolName:
		if len(args) >= 1 {
			input = fmt.Sprintf(`{"file_path": %s}`, strconv.Quote(args[0]))
		} else {
			return UIState{}, fmt.Errorf("o comando 'view' requer um caminho de arquivo")
		}
	}

	response, err := tool.Run(a.ctx, tools.ToolCall{Input: input})
	if err != nil {
		return UIState{}, fmt.Errorf("erro ao executar o comando %s: %w", command, err)
	}

	uiState := a.getCurrentUIState()
	// O método Text() agora existe na ToolResponse e retorna o conteúdo textual.
	uiState.ViewContent = response.Content

	return uiState, nil
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
		ViewContent:   "",
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

// LspClients retorna nil, pois a App da UI não gerencia clientes LSP.
func (a *App) LspClients() *csync.Map[string, *lsp.Client] {
	return a.lspClients
}

// Permissions retorna nil, pois a App da UI não gerencia permissões de ferramentas.
func (a *App) Permissions() permission.Service {
	return a.permissions
}

// Files retorna nil, pois a App da UI não gerencia o histórico de arquivos.
func (a *App) Files() history.Service {
	return a.files
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
