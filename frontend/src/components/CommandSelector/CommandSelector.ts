import './CommandSelector.css';
import { ListDirectory } from '../../../wailsjs/go/api/App';


export interface CommandActions {    
    onUpdateInput: (newText: string) => void;
    getCurrentInput: () => string;
}



export class CommandSelector {
    private element: HTMLElement;
    private visible: boolean = false;
    private selectedIndex: number = -1;
    private currentItems: string[] = [];
    private actions: CommandActions;

    // Comandos disponíveis
    private commands: string[] = [];
    private readonly attachSubCommands = ['add', 'delete'];
    private currentCommand: string | null = null;
    private currentPath: string = '.';

    constructor(actions: CommandActions, commands: string[]) {
        this.actions = actions;
        this.commands = commands;
        this.element = document.createElement('div');
        this.element.className = 'command-selector';
        this.hide();
    }

    setCommands(commands: string[]) {
        this.commands = commands;
    }

    getElement(): HTMLElement {
        return this.element;
    }

    show() {
        this.visible = true;
        this.element.style.display = 'block';
    }

    hide() {
        this.visible = false;
        this.element.style.display = 'none';
        this.currentCommand = null;
        this.currentPath = '.';
        this.selectedIndex = -1;
    }

    isVisible(): boolean {
        return this.visible;
    }

    // Verifica se os itens atuais são subitens (arquivos/subcomandos) ou os comandos principais.
    hasSubItems(): boolean {
        return this.currentCommand !== null;
    }

    hasSelection(): boolean {
        return this.selectedIndex !== -1 && this.currentItems.length > 0;
    }

    async filter(query: string) {
        // Se a busca estiver vazia e nenhum comando estiver ativo, mostra todos os comandos.
        if (query === '' && this.currentCommand === null) {
            this.currentItems = [...this.commands];
            this.updateList(this.currentItems);
            return;
        }

        const queryParts = query.split(' ');
        const commandPart = queryParts[0];
        const subCommandPart = queryParts.length > 1 ? queryParts[1] : '';

        // Lógica para o comando /attach e seus subcomandos
        if (commandPart === 'attach' && !this.attachSubCommands.includes(subCommandPart)) {
            this.currentCommand = 'attach';
            this.currentItems = this.attachSubCommands.filter(sc => sc.startsWith(subCommandPart));
            this.updateList(this.currentItems);
            return;
        }

        // Se um comando já foi selecionado, filtramos os arquivos/pastas
        const pathQuery = queryParts.slice(this.currentCommand === 'attach' ? 2 : 1).join(' ');
        const pathParts = pathQuery.split('/');

        if (this.commands.includes(commandPart)) {
            this.currentCommand = commandPart;
            this.currentPath = pathParts.slice(0, -1).join('/') || '.';
            
            try {
                const fileInfos = await ListDirectory(this.currentPath);
                this.currentItems = fileInfos.map(f => f.isDir ? `${f.name}/` : f.name);
                this.updateList(this.currentItems);
            } catch (error) {
                console.error("Erro ao listar diretório:", error);
                this.currentItems = ['Erro ao carregar...'];
                this.updateList(this.currentItems);
            }
        } else if (this.currentCommand === null) {
            // Filtra a lista de comandos
            this.currentItems = this.commands.filter(c => c.startsWith(query));
            this.updateList(this.currentItems);
        }
    }

    private updateList(items: string[]) {
        this.element.innerHTML = '';
        if (items.length === 0) {
            this.element.innerHTML = '<div class="command-item">Nenhum resultado</div>';
            this.selectedIndex = -1;
            return;
        }

        items.forEach((item, index) => {
            const div = document.createElement('div');
            div.className = 'command-item';
            div.textContent = item;
            div.dataset.index = String(index);
            div.addEventListener('click', () => {
                this.selectedIndex = index;
                this.confirmSelection();
            });
            this.element.appendChild(div);
        });

        this.selectedIndex = 0;
        this.highlightSelectedItem();
    }

    private highlightSelectedItem() {
        this.element.querySelectorAll('.command-item').forEach(el => {
            el.classList.remove('selected');
        });
        const selectedElement = this.element.querySelector(`[data-index="${this.selectedIndex}"]`);
        if (selectedElement) {
            selectedElement.classList.add('selected');
            selectedElement.scrollIntoView({ block: 'nearest' });
        }
    }

    selectNext() {
        if (this.selectedIndex < this.currentItems.length - 1) {
            this.selectedIndex++;
            this.highlightSelectedItem();
        }
    }

    selectPrevious() {
        if (this.selectedIndex > 0) {
            this.selectedIndex--;
            this.highlightSelectedItem();
        }
    }

    confirmSelection() {
        if (this.selectedIndex === -1) return;

        const selectedItem = this.currentItems[this.selectedIndex];

        // Se ainda não selecionamos um comando, apenas o definimos
        if (this.currentCommand === null) {
            this.actions.onUpdateInput(`/${selectedItem} `);
        } else if (this.currentCommand === 'attach' && this.attachSubCommands.includes(selectedItem)) {
            // Se estamos no comando 'attach' e selecionamos 'add' ou 'delete'
            this.actions.onUpdateInput(`/${this.currentCommand} ${selectedItem} `);
        } else {
            if (selectedItem.endsWith('/')) {
                // Se for um diretório, atualiza o input para navegar
                let basePath = `/${this.currentCommand}`;
                const currentInput = this.actions.getCurrentInput();
                const inputParts = currentInput.split(' ');
                if (this.currentCommand === 'attach') {
                    basePath += ` ${inputParts[1]}`;
                }
                const newPath = `${basePath} ${this.currentPath}/${selectedItem}`.replace(/\s+/g, ' ').replace(/\.\/|[/]{2,}/g, '/');
                this.actions.onUpdateInput(newPath);
            } else {
                // Se for um arquivo, completa o caminho no input.
                const currentInput = this.actions.getCurrentInput();
                const inputParts = currentInput.trim().split(' ');
                
                let basePath = `/${this.currentCommand}`;
                if (this.currentCommand === 'attach') {
                    basePath += ` ${inputParts[1]}`; // Mantém o subcomando 'add' ou 'delete'
                }

                const finalPath = filepath.join(this.currentPath, selectedItem);
                const inputText = `${basePath} ${finalPath}`;
                this.actions.onUpdateInput(inputText);
                this.hide(); // Esconde o seletor pois um arquivo foi selecionado.
            }
        }
    }
}

// Simulação de `path.join` para o frontend
const filepath = {
    join: (...args: string[]): string => {
        // Filtra partes vazias que podem surgir de barras duplas
        const parts = args.filter(p => p && p !== '.');
        
        // Lida com caminhos absolutos no início
        let path = parts.join('/');
        
        // Simplifica o caminho (remove '..', '.')
        const newParts: string[] = [];
        for (const part of path.split('/')) {
            if (part === '..') {
                newParts.pop();
            } else if (part !== '.') {
                newParts.push(part);
            }
        }
        return newParts.join('/');
    }
};