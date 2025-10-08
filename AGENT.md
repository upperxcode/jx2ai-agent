# Criando e Integrando Comandos na UI do Agente

Este documento descreve o processo para adicionar um novo comando que pode ser invocado pelo usuário através da interface gráfica (UI) do agente. Usaremos a implementação do comando `/view` como exemplo.

## Visão Geral da Arquitetura

A interação de um comando da UI segue um fluxo específico:

1.  **Frontend (InputBar.ts)**: O usuário digita o comando (ex: `/view meu_arquivo.txt`). O texto completo é enviado para o backend.
2.  **Backend (app.go)**: O método `ExecuteCommand` atua como um **adaptador**. Ele recebe a string, identifica o comando, analisa os argumentos e constrói o `input` JSON estruturado que a ferramenta correspondente espera.
3.  **Backend (Tool)**: A ferramenta (ex: `viewTool`) executa sua lógica usando o `input` JSON. A sua resposta (`ToolResponse`) é capturada pelo `ExecuteCommand`.
4.  **Backend (app.go)**: O `ExecuteCommand` popula uma estrutura `UIState` com o resultado da ferramenta e a retorna para o frontend.
5.  **Frontend (InputBar.ts)**: O frontend recebe o novo `UIState` e atualiza a interface conforme necessário (ex: exibindo o conteúdo de um arquivo no chat).

Este fluxo desacopla a UI (que lida apenas com strings) da lógica das ferramentas (que esperam JSON estruturado).

---

## Passo a Passo: Adicionando o Comando `/view`

Vamos detalhar as modificações necessárias em cada camada.

### 1. Backend - `api/app.go`

Esta é a camada de orquestração principal para os comandos da UI.

#### a. Atualizar o `UIState`

Se o comando precisa retornar dados específicos para a UI, adicionamos um campo ao `UIState`. Para o `/view`, precisávamos de um lugar para colocar o conteúdo do arquivo.

```go
// UIState representa o estado atual da interface que o backend gerencia.
type UIState struct {
	CurrentFile   string   `json:"currentFile"`
	AttachedFiles []string `json:"attachedFiles"`
	ViewContent   string   `json:"viewContent"` // <- Adicionado para o /view
}
```

#### b. Registrar a Ferramenta

Em `NewApp()`, a nova ferramenta é instanciada e adicionada ao `toolbelt` para que o sistema a reconheça.

```go
app.toolbelt = map[string]tools.BaseTool{
    // ... outras ferramentas
    tools.ViewToolName: tools.NewViewTool(nil, nil, ""), // FIXME: Dependências nil
}
```

#### c. Implementar a Lógica do Comando em `ExecuteCommand`

Este é o passo mais importante. Adicionamos um `case` para o novo comando. A responsabilidade aqui é traduzir a string de argumentos do usuário para o formato JSON que a `viewTool.Run()` espera.

```go
func (a *App) ExecuteCommand(commandString string) (UIState, error) {
    // ...
    switch command {
    // ... outros cases
    case tools.ViewToolName:
        if len(args) < 1 {
            return UIState{}, fmt.Errorf("o comando 'view' requer um caminho de arquivo")
        }
        // Constrói o input JSON que a ferramenta `view` espera.
        input = fmt.Sprintf(`{"file_path": %s}`, strconv.Quote(args[0]))
    }

    response, err := tool.Run(a.ctx, tools.ToolCall{Input: input})
    // ...

    // Popula o campo ViewContent no estado da UI com a resposta da ferramenta.
    uiState := a.getCurrentUIState()
    if response != nil {
        uiState.ViewContent = response.Text()
    }
    return uiState, nil
}
```

### 2. Frontend - `frontend/src/components/InputBar/InputBar.ts`

O frontend precisa saber o que fazer com a informação retornada no `UIState`.

No método `executeCommand` do `InputBar`, após receber o `newState` do backend, verificamos se o nosso novo campo `viewContent` contém algo.

```typescript
private async executeCommand(text: string) {
    try {
        const newState: main.UIState = await ExecuteCommand(text);
        this.updateState(newState);

        // Se o comando retornou um conteúdo para ser visualizado,
        // envia como uma mensagem do bot para a janela de chat.
        if (newState.viewContent) {
            this.onSendMessage(newState.viewContent);
        }

        this.inputElement.value = ''; // Limpa o input
        this.commandSelector.hide();
    } catch (error) {
        console.error("Erro ao executar comando:", error);
    }
}
```

### Conclusão

Ao seguir este padrão, adicionar novos comandos que interagem com a UI torna-se uma tarefa sistemática. A chave é usar `app.go` como um adaptador, mantendo o frontend e as ferramentas de backend bem definidos e desacoplados.

---

### Sugestão de Refatoração para Extensibilidade

A abordagem atual em `ExecuteCommand` com um `switch` para cada comando funciona, mas pode se tornar difícil de manter. Uma refatoração pode centralizar a lógica de construção de `input` mais perto de cada ferramenta, tornando `app.go` um orquestrador mais genérico e facilitando a adição de novos comandos sem modificar sua estrutura principal.

O fluxo proposto é:

1.  **Definir uma interface `UICommand`**: No pacote `tools`, criar uma interface que as ferramentas da UI possam implementar.

    ```go
    // api/internal/llm/tools/tools.go
    
    // UICommand é uma interface que ferramentas podem implementar para construir
    // seu input a partir de argumentos de string da UI.
    type UICommand interface {
    	BuildInput(args []string) (string, error)
    }
    ```

2.  **Implementar `UICommand` nas Ferramentas**: Cada ferramenta que pode ser chamada pela UI (como `view`, `write`, etc.) implementará esta interface. A lógica de converter `[]string` para o JSON de `input` é movida para dentro da própria ferramenta.

    ```go
    // Exemplo para a viewTool
    // api/internal/llm/tools/view.go
    
    // BuildInput implementa a interface UICommand para a viewTool.
    func (v *viewTool) BuildInput(args []string) (string, error) {
    	if len(args) < 1 {
    		return "", fmt.Errorf("o comando 'view' requer um caminho de arquivo")
    	}
    	// Constrói o input JSON que a ferramenta `view` espera.
    	input := fmt.Sprintf(`{"file_path": %s}`, strconv.Quote(args[0]))
    	return input, nil
    }
    ```

3.  **Simplificar `ExecuteCommand`**: O método se torna um orquestrador genérico, que apenas delega a construção do `input` para a ferramenta correta, sem precisar conhecer os detalhes de cada uma.

    ```go
    // api/app.go
    func (a *App) ExecuteCommand(commandString string) (UIState, error) {
        // ... (análise do comando e argumentos)
    
        tool, ok := a.toolbelt[command]
        if !ok { /* ... */ }
    
        // Verifica se a ferramenta implementa a interface para construir o input.
        uiTool, ok := tool.(tools.UICommand)
        if !ok {
            return UIState{}, fmt.Errorf("o comando '%s' não está configurado para ser usado pela UI", command)
        }
    
        // Delega a construção do input para a ferramenta.
        input, err := uiTool.BuildInput(args)
        if err != nil {
            return UIState{}, fmt.Errorf("erro ao construir input para o comando %s: %w", command, err)
        }
    
        response, err := tool.Run(a.ctx, tools.ToolCall{Input: input})
        // ... (processamento da resposta)
    }
    ```

#### Vantagens

*   **Melhor Coesão:** A lógica para construir o `input` de uma ferramenta reside com a própria ferramenta.
*   **Menor Acoplamento:** `app.go` não precisa mais conhecer os detalhes dos argumentos de cada comando.
*   **Mais Extensível:** Para adicionar um novo comando à UI, basta criar a nova ferramenta, implementar as interfaces `BaseTool` e `UICommand`, e registrá-la. Nenhuma modificação em `app.go` é necessária.