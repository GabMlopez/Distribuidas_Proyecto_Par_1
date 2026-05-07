import './globals.css';
import { UserProvider } from '@/context/UserContext';
import { ReactNode } from 'react';

export const metadata = {
  title: 'Chat Distribuido',
  description: 'Aplicación de Chat Distribuido',
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="es">
      <body>
        <div id="app" className="app-container">
          <UserProvider>
            {children}
          </UserProvider>
        </div>
      </body>
    </html>
  );
}
