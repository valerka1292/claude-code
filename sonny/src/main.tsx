import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import App from './App.tsx';
import './index.css';
import { ProvidersProvider } from './context/ProvidersContext';

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <ProvidersProvider>
      <App />
    </ProvidersProvider>
  </StrictMode>,
);
