import { useState } from 'react';
import { Scene } from './components/Scene';
import { LoginScreen } from './components/LoginScreen';
import { Dashboard } from './components/Dashboard';
import type { Credentials } from './api/client';
import './index.css';

function App() {
  const [creds, setCreds] = useState<Credentials | null>(null);

  const handleLogin = (credentials: Credentials) => {
    setCreds(credentials);
  };

  const handleLogout = () => {
    setCreds(null);
  };

  return (
    <Scene>
      {creds ? (
        <Dashboard creds={creds} onLogout={handleLogout} />
      ) : (
        <LoginScreen onLogin={handleLogin} />
      )}
    </Scene>
  );
}

export default App;
