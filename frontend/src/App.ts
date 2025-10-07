import { ChatWindow } from './components/ChatWindow/ChatWindow';
import { InputBar } from './components/InputBar/InputBar';

export class App {
    private element: HTMLElement;
    private chatWindow: ChatWindow;
    private inputBar: InputBar;

    constructor(element: HTMLElement) {
        this.element = element;

        // 1. Cria a estrutura principal do layout
        // O .chatbot-container usará flexbox para organizar os filhos.
        this.element.innerHTML = `
            <div class="chatbot-container">
                <div class="chat-window-wrapper"></div>
                <div class="input-bar-wrapper"></div>
            </div>
        `;

        // 2. Instancia os componentes nos seus respectivos contêineres
        const chatWindowWrapper = this.element.querySelector('.chat-window-wrapper') as HTMLElement;
        const inputBarWrapper = this.element.querySelector('.input-bar-wrapper') as HTMLElement;
        this.chatWindow = new ChatWindow(chatWindowWrapper);
        this.inputBar = new InputBar(inputBarWrapper, (message) => {
            this.handleSendMessage(message);
        });
    }

    private handleSendMessage(message: string) {
        this.chatWindow.addMessage(message, 'user');
    }
}