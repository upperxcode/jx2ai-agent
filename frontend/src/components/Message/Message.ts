import './Message.css';

type MessageProps = {
    text: string;
    sender: 'user' | 'bot';
}

export class Message {
    private props: MessageProps;

    constructor(props: MessageProps) {
        this.props = props;
    }

    render(): HTMLElement {
        const messageElement = document.createElement('div');
        messageElement.className = `message ${this.props.sender}`;
        messageElement.textContent = this.props.text;
        return messageElement;
    }
}