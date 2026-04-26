import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import App from './App.tsx';
import './index.css';
import { ProvidersProvider } from './context/ProvidersContext';
import { StorageProvider } from './context/StorageContext';

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <StorageProvider>
      <ProvidersProvider>
        <App />
      </ProvidersProvider>
    </StorageProvider>
  </StrictMode>,
);
