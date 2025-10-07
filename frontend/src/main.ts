import { App } from './App';

const appContainer = document.querySelector('#app')! as HTMLElement;

// Cria a instância inicial da aplicação
let app = new App(appContainer);

// Lógica de Hot Module Replacement (HMR)
if (import.meta.hot) {
    import.meta.hot.accept(() => {
        console.log('HMR: Recriando a aplicação...');
        // Limpa o container da aplicação antiga
        appContainer.innerHTML = '';
        // Cria uma nova instância da aplicação com o código atualizado
        app = new App(appContainer);
    });
}
