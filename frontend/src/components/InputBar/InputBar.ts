import './InputBar.css';
import {
    CommandSelector,
    CommandActions,
} from '../CommandSelector/CommandSelector';
// Importe o runtime do Wails para ter acesso às funções do Go

import {
    ReadFile,
    ExecuteCommand,
    ListCommands,
} from '../../../wailsjs/go/api/App';
import { main } from '../../../wailsjs/go/models';

export class InputBar {
    private element: HTMLElement;
    private onSendMessage: (message: string) => void = () => {};

    private inputElement: HTMLInputElement;
    private commandSelector: CommandSelector;

    // Estado
    private currentFile: string | null = null;
    private isCurrentFileActive: boolean = true;
    private attachedFiles: string[] = [];

    constructor(container: HTMLElement, onSendMessage: (message: string) => void) {
        this.element = container;
        const commandActions: CommandActions = {
            onUpdateInput: this.updateInputValue.bind(this),
            getCurrentInput: () => this.inputElement.value,
        };
        this.commandSelector = new CommandSelector(commandActions, []);

        this.onSendMessage = onSendMessage;

        // 1. Cria a estrutura HTML inicial
        this.element.className = 'input-area-container';
        this.element.innerHTML = this.getHTML();

        // 2. Obtém as referências dos elementos DOM
        this.inputElement = document.createElement('input');
        this.inputElement.className = 'chat-input';
        this.inputElement.type = 'text';
        this.inputElement.placeholder = 'Digite uma mensagem ou / para comandos...';

        const inputWrapper = this.element.querySelector('.input-wrapper')!;
        inputWrapper.prepend(this.commandSelector.getElement());
        inputWrapper.prepend(this.inputElement);
        this.setupEventListeners();
        this.loadCommands();
    }

    private async loadCommands() {
        const commands = await ListCommands();
        const commandNames = commands.map((c: main.CommandInfo) => c.name);
        this.commandSelector.setCommands(commandNames);
    }

    private getHTML(): string {
        // Ícone de arquivo (SVG inline para simplicidade)
        const fileIcon = `
            <svg viewBox="0 0 24 24">
                <path d="M14,2H6A2,2 0 0,0 4,4V20A2,2 0 0,0 6,22H18A2,2 0 0,0 20,20V8L14,2M18,20H6V4H13V9H18V20Z" />
            </svg>
        `;

        // Ícone de envio (SVG inline)
        const sendIcon = `
            <svg viewBox="0 0 24 24">
                <path d="M2,21L23,12L2,3V10L17,12L2,14V21Z" />
            </svg>
        `;

        // Ícone de anexo (clips)
        const attachIcon = `
            <svg viewBox="0 0 24 24">
                <path d="M16.5,6V17.5A4,4 0 0,1 12.5,21.5A4,4 0 0,1 8.5,17.5V5A2.5,2.5 0 0,1 11,2.5A2.5,2.5 0 0,1 13.5,5V15.5A1,1 0 0,1 12.5,16.5A1,1 0 0,1 11.5,15.5V6H10V15.5A2.5,2.5 0 0,0 12.5,18A2.5,2.5 0 0,0 15,15.5V5A4,4 0 0,0 11,1A4,4 0 0,0 7,5V17.5A5.5,5.5 0 0,0 12.5,23A5.5,5.5 0 0,0 18,17.5V6H16.5Z" />
            </svg>
        `;

        // Ícone de lixeira
        const trashIcon = `
            <svg viewBox="0 0 24 24">
                <path d="M9,3V4H4V6H5V19A2,2 0 0,0 7,21H17A2,2 0 0,0 19,19V6H20V4H15V3H9M7,6H17V19H7V6M9,8V17H11V8H9M13,8V17H15V8H13Z" />
            </svg>
        `;

        return `
            <div class="icon-bar">
                <div class="icon-item">
                    <input type="checkbox" id="current-file-checkbox" ${this.isCurrentFileActive ? 'checked' : ''}>
                     ${fileIcon}
                    <span class="file-name">${this.currentFile || 'Nenhum'}</span>
                </div>
                <div class="icon-item attach-icon-wrapper">
                    ${attachIcon}
                    <span>${this.attachedFiles.length}</span>
                    <div class="attached-files-popup">
                        <div class="attached-files-list">
                            <!-- Arquivos anexados serão inseridos aqui -->
                        </div>
                        <div class="clear-attachments-wrapper" title="Limpar todos os anexos">${trashIcon}</div>
                    </div>
                </div>
            </div>
            <div class="input-wrapper">
                <span class="send-icon-wrapper">${sendIcon}</span>
            </div>
        `;
    }

    private render() {
        // Este método agora apenas atualiza as partes dinâmicas da UI
        const fileNameSpan = this.element.querySelector('.file-name');
        if (fileNameSpan) {
            fileNameSpan.textContent = this.currentFile || 'Nenhum';
        }

        const checkbox = this.element.querySelector('#current-file-checkbox') as HTMLInputElement;
        if (checkbox) {
            checkbox.checked = this.isCurrentFileActive;
        }

        const attachedList = this.element.querySelector('.attached-files-list');
        if (attachedList) {
            attachedList.innerHTML = this.attachedFiles.length > 0
                ? this.attachedFiles.map(f => `<div class="attached-file-item">${f}</div>`).join('')
                : '<span>Nenhum arquivo anexado.</span>';
        }

        const attachCount = this.element.querySelector('.attach-icon-wrapper span');
        if (attachCount) {
            attachCount.textContent = String(this.attachedFiles.length);
        }
    }

    private updateState(newState: main.UIState) {
        this.currentFile = newState.currentFile || null;
        this.attachedFiles = newState.attachedFiles || [];
        this.render();
    }

    private setupEventListeners() {
        // O inputElement agora é criado uma vez no construtor, então podemos adicionar os listeners aqui
        this.inputElement.addEventListener('input', this.handleInput.bind(this));
        this.inputElement.addEventListener('keydown', this.handleKeyDown.bind(this));

        // Listener para o checkbox
        this.element.addEventListener('change', (e) => {
            const target = e.target as HTMLInputElement;
            if (target.id === 'current-file-checkbox') {
                this.isCurrentFileActive = target.checked;
                console.log('Inclusão do arquivo atual:', this.isCurrentFileActive);
            }
        });

        this.element.querySelector('.clear-attachments-wrapper')?.addEventListener('click', (e) => {
            e.stopPropagation(); // Impede que o clique feche o popup imediatamente
            this.attachedFiles = [];
            this.render();
        });

        const attachWrapper = this.element.querySelector('.attach-icon-wrapper');
        const attachPopup = this.element.querySelector('.attached-files-popup') as HTMLElement;

        attachWrapper?.addEventListener('click', (e) => {
            e.stopPropagation(); // Impede que o clique se propague para o document
            attachPopup.classList.toggle('show');
        });

        document.addEventListener('click', () => {
            attachPopup.classList.remove('show');
        });

        this.element.querySelector('.send-icon-wrapper')?.addEventListener('click', () => {
            this.sendMessage();
        });


    }

    private updateInputValue(newText: string) {
        this.inputElement.value = newText;
        // Dispara o evento de input manualmente para re-filtrar a lista de comandos/arquivos
        this.inputElement.dispatchEvent(new Event('input'));
        // Coloca o foco de volta no input
        this.inputElement.focus();
    }

    private handleInput() {
        const text = this.inputElement.value;
        if (text.startsWith('/')) {
            const query = text.substring(1);
            this.commandSelector.show();
            this.commandSelector.filter(query);
        } else {
            this.commandSelector.hide();
        }
    }

    private handleKeyDown(e: KeyboardEvent) {
        if (this.commandSelector.isVisible()) {
            switch (e.key) {
                case 'ArrowUp':
                    e.preventDefault();
                    this.commandSelector.selectPrevious();
                    break;
                case 'ArrowDown':
                    e.preventDefault();
                    this.commandSelector.selectNext();
                    break;
                case 'Enter':
                    e.preventDefault();
                    // Se a lista está visível, primeiro usamos o Enter para confirmar a seleção,
                    // o que preenche o input. A execução real acontecerá no próximo Enter.
                    // Se o usuário já digitou um comando completo, executamos diretamente.
                    const text = this.inputElement.value.trim();
                    const parts = text.substring(1).split(' ');
                    const isCompleteCommand = parts.length > 1 && (parts[0] === 'currentfile' || (parts[0] === 'attach' && parts.length > 2));

                    if (isCompleteCommand) {
                        this.executeCurrentInput();
                    } else if (this.commandSelector.hasSelection()) {
                        this.commandSelector.confirmSelection();
                    }

                    break;
                case 'Escape':
                    e.preventDefault();
                    this.commandSelector.hide();
                    break;
            }
        } else if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            this.executeCurrentInput();
        }
    }

    private executeCurrentInput() {
        const text = this.inputElement.value.trim();
        if (text.startsWith('/')) {
            console.log("Executando comando:", text);
            this.executeCommand(text);
        } else if (text) {
            console.log("Enviando mensagem:", text);
            this.sendMessage();
        }
    }

    private async executeCommand(text: string) {
        try {
            const newState = await ExecuteCommand(text);
            this.updateState(newState);
            this.inputElement.value = ''; // Limpa o input após o sucesso
            this.commandSelector.hide();
        } catch (error) {
            console.error("Erro ao executar comando:", error);
            // Opcional: mostrar um erro para o usuário
        }
    }

    private async sendMessage() {
        const message = this.inputElement.value.trim();
        if (!message) return;

        let finalMessage = message;

        // Constrói o contexto a partir dos arquivos
        let context = '';
        if (this.isCurrentFileActive && this.currentFile) {
            context += `Arquivo Atual: "${this.currentFile}"\n\n`;
        }
        if (this.attachedFiles.length > 0) {
            context += `Arquivos Anexados: ${this.attachedFiles.join(', ')}\n\n`;
        }

        if (context) {
            try {
                const filesToRead = this.isCurrentFileActive && this.currentFile ? [this.currentFile, ...this.attachedFiles] : [...this.attachedFiles];
                const uniqueFiles = [...new Set(filesToRead)]; // Evita ler o mesmo arquivo duas vezes

                let contentBlocks = '';
                for (const file of uniqueFiles) {
                    const fileContent = await ReadFile(file);
                    contentBlocks += `--- Conteúdo de ${file} ---\n\`\`\`\n${fileContent}\n\`\`\`\n\n`;
                }
                finalMessage = `Com base nos arquivos abaixo:\n\n${contentBlocks}\n\n${message}`;
            } catch (error) {
                console.error("Erro ao ler o arquivo de contexto:", error);
                // Opcional: informar o usuário sobre o erro
                finalMessage = `(Erro ao ler o arquivo ${this.currentFile}) ${message}`;
            }
        }

        this.onSendMessage(finalMessage);
        this.inputElement.value = '';
    }
}