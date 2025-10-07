import { Message } from '../Message/Message';
import './ChatWindow.css';

export class ChatWindow {
    private element: HTMLElement;

    constructor(element: HTMLElement) {
        this.element = element;
        this.element.className = 'chat-window';
        this.addMessage("Olá! Como posso ajudar você hoje?", 'bot');
    }

    public addMessage(text: string, sender: 'user' | 'bot') {
        const message = new Message({ text, sender });
        this.element.appendChild(message.render());
        this.scrollToBottom();
    }

    private scrollToBottom() {
        this.element.scrollTop = this.element.scrollHeight;
    }
}